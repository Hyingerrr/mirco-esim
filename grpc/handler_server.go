package grpc

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/jukylin/esim/container"

	tracerid "github.com/jukylin/esim/pkg/tracer-id"
	"github.com/jukylin/esim/pkg/validate"
	opentracing2 "github.com/opentracing/opentracing-go"

	logx "github.com/jukylin/esim/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/davecgh/go-spew/spew"
	"google.golang.org/grpc/metadata"

	"github.com/jukylin/esim/core/meta"

	"google.golang.org/grpc"
)

var checker = validate.NewValidateRepo()

func (gs *Server) handleServer() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			start       = time.Now()
			cancel      func()
			serviceName = container.GetServiceName()
		)

		// get timeout from ctx, compare with the config
		timeout := gs.config.Timeout
		if dl, ok := ctx.Deadline(); ok {
			if out := time.Until(dl); timeout > out {
				timeout = out
			}
		}
		ctx, cancel = context.WithTimeout(ctx, timeout*time.Millisecond)
		defer cancel()

		md := meta.MD{}
		if mdCtx, ok := metadata.FromIncomingContext(ctx); ok {
			for k, v := range mdCtx {
				if _, o := md[k]; !o {
					md[k] = v[0]
				}
			}
		}
		ctx = meta.NewContext(ctx, md)

		// request logger
		if gs.config.Debug {
			logx.Infoc(ctx, "Get_Request_Params: method[%v], server[%v], ctxParams:%+v, body:%+v",
				info.FullMethod, info.Server, md.Marshal(), req)
		}

		// monitor
		if gs.config.Metrics {
			// debug
			var appId string
			if appId = meta.String(ctx, meta.AppID); appId == "" {
				logx.Errorc(ctx, "AppID_Is_Empty: method[%v], appId[%v]", info.FullMethod, meta.String(ctx, meta.AppID))
			}
			_serverGRPCReqQPS.Inc(serviceName, info.FullMethod, meta.String(ctx, meta.AppID))
		}

		resp, err = handler(ctx, req)

		// response logger
		if gs.config.Debug {
			logx.Infoc(ctx, "Send_Response_Params: method[%v], server[%v], cost[%v], resp_params:%+v",
				info.FullMethod, info.Server, time.Since(start).String(), spew.Sdump(resp))
		}

		// monitor
		if gs.config.Metrics {
			_serverGRPCReqDuration.Observe(float64(time.Since(start)/time.Millisecond),
				serviceName, info.FullMethod, meta.String(ctx, meta.AppID))
		}

		// check grpc slow
		if slowTime := gs.config.SlowTime; slowTime != 0 {
			if sub := time.Now().Sub(start); sub > slowTime*time.Millisecond {
				logx.Warnc(ctx, "Slow server %s, cost:%v", info.FullMethod, sub)
			}
		}

		return resp, err
	}
}

func (gs *Server) recovery() grpc.UnaryServerInterceptor {
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
func (gs *Server) tracerID() grpc.UnaryServerInterceptor {
	tracerID := tracerid.TracerID()
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		sp := opentracing2.SpanFromContext(ctx)
		if sp == nil {
			ctx = context.WithValue(ctx, tracerid.ActiveEsimKey, tracerID())
		}
		resp, err = handler(ctx, req)
		return resp, err
	}
}

func (gs *Server) validate() grpc.UnaryServerInterceptor {
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
