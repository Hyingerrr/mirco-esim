package grpc

import (
	"context"
	"fmt"
	"runtime"
	"time"

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

// ------------------------------------------------------ //
// --------------------  server  ----------------------- //
// ------------------------------------------------------ //
func (gs *Server) metaServerData() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var (
			start  = time.Now()
			cancel func()
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

		cmd := meta.MD{}
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			for k, v := range md {
				if _, o := cmd[k]; !o {
					cmd[k] = v[0]
				}
			}
		}

		ctx = meta.NewContext(ctx, cmd)

		// request logger
		if gs.config.Debug {
			logx.Infoc(ctx, "Get_Request_Params: method[%v], server[%v], ctxParams:%+v, body:%+v",
				info.FullMethod, info.Server, cmd.Marshal(), req)
		}

		// monitor
		if gs.config.Metrics {
			_serverGRPCReqQPS.Inc(meta.String(ctx, meta.ServiceName), info.FullMethod, meta.String(ctx, meta.AppID))
		}

		fmt.Println(23)
		resp, err = handler(ctx, req)
		fmt.Println(24)
		duration := time.Since(start)
		// response logger
		if gs.config.Debug {
			logx.Infoc(ctx, "Send_Response_Params: method[%v], server[%v], cost[%v], resp_params:%+v",
				info.FullMethod, info.Server, duration.String(), spew.Sdump(resp))
		}

		// monitor
		if gs.config.Metrics {
			_serverGRPCReqDuration.Observe(float64(duration/time.Millisecond),
				meta.String(ctx, meta.ServiceName), info.FullMethod, meta.String(ctx, meta.AppID))
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

		fmt.Println(21)
		resp, err = handler(ctx, req)
		fmt.Println(22)
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
		fmt.Println(27)
		resp, err = handler(ctx, req)
		fmt.Println(28)
		return resp, err
	}
}

func (gs *Server) validate() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if err = checker.ValidateStruct(req); err != nil {
			err = status.Error(codes.InvalidArgument, err.Error())
			return
		}
		fmt.Println(25)
		resp, err = handler(ctx, req)
		fmt.Println(26)
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
