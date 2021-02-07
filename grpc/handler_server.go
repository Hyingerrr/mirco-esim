package grpc

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/jukylin/esim/core/tracer"
	"github.com/opentracing/opentracing-go/ext"
	opentracinglog "github.com/opentracing/opentracing-go/log"

	"github.com/jukylin/esim/container"

	tracerid "github.com/jukylin/esim/pkg/tracer-id"
	"github.com/jukylin/esim/pkg/validate"
	"github.com/opentracing/opentracing-go"

	logx "github.com/jukylin/esim/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/metadata"

	"github.com/jukylin/esim/core/meta"

	"google.golang.org/grpc"
)

var checker = validate.NewValidateRepo()

func timeoutUnaryServerInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var cancel func()

		// get timeout from ctx, compare with the config
		if dl, ok := ctx.Deadline(); ok {
			if out := time.Until(dl); timeout > out {
				timeout = out
			}
		}

		// debug
		logx.Infoc(ctx, "Server Deadline timeOut:%v", timeout)
		ctx, cancel = context.WithTimeout(ctx, timeout*time.Millisecond)
		defer cancel()
		dls, _ := ctx.Deadline()
		logx.Infoc(ctx, "Server Deadline:%v", dls)

		resp, err = handler(ctx, req)

		return resp, err
	}
}

func debugUnaryServerInterceptor(slowTime time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			start = time.Now()
		)
		// request logger
		logx.Infoc(ctx, "Get_Request_Params: method[%v], server[%v], body:%+v",
			info.FullMethod, info.Server, req)

		resp, err = handler(ctx, req)

		// response logger
		logx.Infoc(ctx, "Send_Response_Params: method[%v], server[%v], cost[%v], resp_params:%+v",
			info.FullMethod, info.Server, time.Since(start).String(), spew.Sdump(resp))

		// check grpc slow
		if sub := time.Now().Sub(start); sub > slowTime*time.Millisecond {
			logx.Warnc(ctx, "Slow server %s, cost:%v", info.FullMethod, sub)
		}

		return resp, err
	}
}

func metricUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var (
		start       = time.Now()
		serviceName = container.AppName()
		appId       string
	)

	if mdCtx, ok := metadata.FromIncomingContext(ctx); ok {
		if ai := mdCtx.Get(meta.AppID); len(ai) > 0 {
			appId = ai[0]
		} else {
			logx.Errorc(ctx, "AppID_Is_Empty: method[%v], appId[%v]", info.FullMethod, appId)
		}
	}

	// monitor
	_serverGRPCReqQPS.Inc(serviceName, info.FullMethod, appId)

	resp, err := handler(ctx, req)

	// monitor
	_serverGRPCReqDuration.Observe(float64(time.Since(start)/time.Millisecond),
		serviceName, info.FullMethod, appId)

	return resp, err
}

func traceUnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	esimTracer := opentracing.GlobalTracer()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	spCtx, err := esimTracer.Extract(opentracing.TextMap, tracer.MetadataReaderWriter{MD: md})
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		// todo something
		logx.Errorc(ctx, "extract from metadata err:%v", err)
		return handler(ctx, req)
	}

	span := esimTracer.StartSpan(
		info.FullMethod,
		ext.RPCServerOption(spCtx),
		tracer.TagComponent("gRPC"),
		ext.SpanKindRPCServer,
		tracer.TagLocalIPV4(),
	)
	defer span.Finish()

	newCtx := opentracing.ContextWithSpan(ctx, span)
	resp, err := handler(newCtx, req)
	if err != nil {
		code := codes.Unknown
		if s, ok := status.FromError(err); ok {
			code = s.Code()
		}
		span.SetTag("gRPC_code", code)
		ext.Error.Set(span, true)
		span.LogFields(opentracinglog.String("event", "error"), opentracinglog.String("message", err.Error()))
	}

	span.SetTag("gRPC_code", codes.OK)

	return resp, err
}

func recoverServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = recoverFrom(r, info.FullMethod)
			}
		}()
		resp, err = handler(ctx, req)
		return
	}
}

// tracerId If not found opentracing's tracer_id then generate a new tracer_id.
// Recommend to the end of the Interceptor.
func tracerIDServerInterceptor() grpc.UnaryServerInterceptor {
	tracerID := tracerid.TracerID()
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		sp := opentracing.SpanFromContext(ctx)
		if sp == nil {
			ctx = context.WithValue(ctx, tracerid.ActiveEsimKey, tracerID())
		}
		resp, err = handler(ctx, req)
		return resp, err
	}
}

func validateServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if err = checker.ValidateStruct(req); err != nil {
			err = status.Error(codes.InvalidArgument, err.Error())
			return
		}
		resp, err = handler(ctx, req)
		return
	}
}

func recoverFrom(r interface{}, fullMethod string) error {
	var stacktrace string
	for i := 1; i < 4; i++ {
		_, f, l, got := runtime.Caller(i)
		if !got {
			break
		}

		stacktrace += fmt.Sprintf("%s:%d\n", f, l)
	}

	logx.Errorf("Handle_GRPC_PANIC. fullMethod[%v], panic: %v, stack: %v",
		fullMethod, r, stacktrace)

	return status.Errorf(codes.Unknown, "%s", r)
}

func extractSpanContext(ctx context.Context) (opentracing.SpanContext, error) {
	return nil, nil
}
