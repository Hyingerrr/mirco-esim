package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jukylin/esim/core/rpcode"

	"github.com/jukylin/esim/container"

	"github.com/jukylin/esim/core/tracer"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	opentracinglog "github.com/opentracing/opentracing-go/log"
	"google.golang.org/grpc/codes"

	"github.com/pkg/errors"
	"google.golang.org/grpc/status"

	"github.com/davecgh/go-spew/spew"
	"github.com/jukylin/esim/config"
	logx "github.com/jukylin/esim/log"

	"google.golang.org/grpc/metadata"

	"github.com/jukylin/esim/core/meta"

	"google.golang.org/grpc"
)

func timeOutUnaryClientInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var (
			timeOpt *TimeoutCallOption
			cancel  context.CancelFunc
		)

		// request timeout ctrl
		// timeout ctx must before the all
		for _, opt := range opts {
			var ok bool
			timeOpt, ok = opt.(*TimeoutCallOption)
			if ok {
				break
			}
		}
		if timeOpt != nil && timeOpt.Timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeOpt.Timeout/1000)
		} else {
			ctx, cancel = context.WithTimeout(ctx, timeout/1000)
		}
		defer cancel()

		//debug
		dl, _ := ctx.Deadline()
		logx.Infoc(ctx, "request deadline: %v", dl.String())

		err := invoker(ctx, method, req, reply, cc, opts...)

		return handlerErr(err)
	}
}

func metadataHandler() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// set metadata
		md, err := setClientMetadata(ctx, req)
		if err != nil {
			return handlerErr(err)
		}

		if len(md) > 0 {
			ctx = metadata.NewOutgoingContext(ctx, md)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func metricUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var (
			beg        = time.Now()
			md         metadata.MD
			serverName = container.AppName()
		)

		// merge with old matadata if exists
		if oldmd, ok := metadata.FromOutgoingContext(ctx); ok {
			md = metadata.Join(md, oldmd)
		}
		ctx = metadata.NewOutgoingContext(ctx, md)

		logx.Infoc(ctx, "Set_Client_Metadata: %v", md)

		err := invoker(ctx, method, req, reply, cc, opts...)
		rpcStatus := rpcode.ExtractCode(err)

		var getAppID = func() string {
			if ai := md.Get(meta.AppID); len(ai) > 0 {
				return ai[0]
			}
			return ""
		}

		// metrics
		_clientGRPCReqDuration.Observe(float64(time.Since(beg)/time.Millisecond), serverName, method, getAppID())
		_clientGRPCReqQPS.Inc(serverName, method, getAppID(), rpcStatus.Code)

		return handlerErr(err)
	}
}

func debugUnaryClientInterceptor(slowTime time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		beg := time.Now()

		// request params
		logx.Infoc(ctx, "GRPC_Send_RequestParams: method[%v], target[%v], params:%+v",
			method, cc.Target(), req)

		err := invoker(ctx, method, req, reply, cc, opts...)

		// check slow
		if sub := time.Now().Sub(beg); sub > slowTime {
			logx.Warnc(ctx, "Slow_Client_GRPC handle: %s, target: %v, cost: %v", method, cc.Target(), sub)
		}

		// response params
		logx.Infoc(ctx, "GRPC_Get_ResponseParams: method[%v], target[%v], params:%+v",
			method, cc.Target(), spew.Sdump(reply))

		return handlerErr(err)
	}
}

func traceUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		span, ctx := opentracing.StartSpanFromContext(
			ctx,
			method,
			tracer.TagComponent("gRPC"),
			tracer.TagLocalIPV4(),
			ext.SpanKindRPCClient,
		)
		defer span.Finish()

		newCtx := injectSpanContext(ctx, span)
		err := invoker(newCtx, method, req, reply, cc, opts...)
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
		return err
	}
}

func injectSpanContext(ctx context.Context, span opentracing.Span) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		// 对metadata进行修改，需要用拷贝的副本进行修改
		md = md.Copy()
	}

	carrier := tracer.MetadataReaderWriter{MD: md}
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.TextMap, carrier)
	if err != nil {
		span.LogFields(opentracinglog.String("event", "Tracer.Inject() failed"), opentracinglog.Error(err))
	}

	return metadata.NewOutgoingContext(ctx, md)
}

func setClientMetadata(ctx context.Context, req interface{}) (metadata.MD, error) {
	var (
		cmd     = new(meta.CommonHeader)
		d       = metadata.MD{}
		marshal = func(v interface{}) []byte {
			b, _ := json.Marshal(v)
			return b
		}
	)

	if err := json.Unmarshal(marshal(req), cmd); err != nil {
		logx.Errorc(ctx, "Metadata_Unmarshal err: %v", err)
		return nil, err
	}

	if cmd.Head == nil {
		return d, nil
	}

	d.Set(meta.AppID, cmd.Head.AppId)
	d.Set(meta.TermNO, cmd.Head.TermNo)
	d.Set(meta.MerID, cmd.Head.MerchNo)
	d.Set(meta.ProdCd, cmd.Head.ProdCd)
	d.Set(meta.TranCd, cmd.Head.TranCd)
	d.Set(meta.TranSeq, cmd.Head.TranSeq)
	d.Set(meta.Protocol, meta.RPCProtocol)

	if srcSysId := cmd.Head.SrcSysId; srcSysId == "" {
		d.Set(meta.SrcSysId, config.GetString("appname"))
	} else {
		d.Set(meta.SrcSysId, srcSysId)
	}

	// 可以根据服务发现，自动获取目的服务名
	if dstSysId := cmd.Head.DstSysId; dstSysId == "" {
		d.Set(meta.DstSysId, config.GetString("appname"))
	} else {
		d.Set(meta.DstSysId, dstSysId)
	}

	if traceID := cmd.Head.TraceId; traceID == "" {
		d.Set(meta.TraceID, fmt.Sprintf("%v", time.Now().UnixNano()))
	} else {
		d.Set(meta.TraceID, cmd.Head.TraceId)
	}

	return d, nil
}

func handlerErr(err error) error {
	gst, ok := status.FromError(err)
	if ok {
		return errors.WithMessage(err, gst.Message())
	}
	return err
}
