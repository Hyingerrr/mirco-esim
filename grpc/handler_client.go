package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/status"

	"github.com/davecgh/go-spew/spew"
	"github.com/jukylin/esim/config"
	logx "github.com/jukylin/esim/log"

	"google.golang.org/grpc/metadata"

	"github.com/jukylin/esim/core/meta"

	"google.golang.org/grpc"
)

func (gc *ClientOptions) handleClient() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var beg = time.Now()
		var codes = "200"

		// request timeout ctrl
		// timeout ctx must before the all
		var timeOpt *TimeoutCallOption
		var cancel context.CancelFunc
		for _, opt := range opts {
			var ok bool
			timeOpt, ok = opt.(*TimeoutCallOption)
			if ok {
				break
			}
		}
		if timeOpt != nil && timeOpt.Timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeOpt.Timeout)
		} else {
			ctx, cancel = context.WithTimeout(ctx, gc.config.Timeout)
		}
		defer cancel()

		// set metadata
		md, err := setClientMetadata(req)
		if err != nil {
			codes = "406"
			return handlerErr(err)
		}

		// merge with old matadata if exists
		if oldmd, ok := metadata.FromOutgoingContext(ctx); ok {
			md = metadata.Join(md, oldmd)
		}
		ctx = metadata.NewOutgoingContext(ctx, md)

		logx.Infoc(ctx, "Set_Client_Metadata: %v", md)

		err = invoker(ctx, method, req, reply, cc, opts...)

		// metrics
		if gc.config.Metrics {
			var serverName, appId string
			if sn := md.Get(meta.ServiceName); len(sn) > 0 {
				serverName = sn[0]
			}
			if ai := md.Get(meta.AppID); len(ai) > 0 {
				appId = ai[0]
			}
			_clientGRPCReqDuration.Observe(float64(time.Since(beg)/time.Millisecond), serverName, method, appId)
			_clientGRPCReqQPS.Inc(serverName, method, appId, codes)
		}

		return handlerErr(err)
	}
}

func (gc *ClientOptions) addClientDebug() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		beg := time.Now()

		// request params
		if gc.config.Debug {
			logx.Infoc(ctx, "GRPC_Send_RequestParams: method[%v], target[%v], params:%+v",
				method, cc.Target(), spew.Sdump(req))
		}

		err := invoker(ctx, method, req, reply, cc, opts...)

		// check slow
		if gc.config.SlowTime != 0 {
			if sub := time.Now().Sub(beg); sub > gc.config.SlowTime {
				logx.Warnc(ctx, "Slow_Client_GRPC handle: %s, target: %v, cost: %v", method, cc.Target(), sub)
			}
		}

		// response params
		if gc.config.Debug {
			logx.Infoc(ctx, "GRPC_Get_ResponseParams: method[%v], target[%v], params:%+v",
				method, cc.Target(), spew.Sdump(reply))
		}

		return handlerErr(err)
	}
}

// client
func setClientMetadata(req interface{}) (metadata.MD, error) {
	var (
		cmd     = new(meta.CommonHeader)
		d       = metadata.MD{}
		marshal = func(v interface{}) []byte {
			b, _ := json.Marshal(v)
			return b
		}
	)

	if err := json.Unmarshal(marshal(req), cmd); err != nil {
		logx.Errorf("Metadata_Unmarshal err: %v", err)
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
	d.Set(meta.ServiceName, config.GetString("appname"))
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
