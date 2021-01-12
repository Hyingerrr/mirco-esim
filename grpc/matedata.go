package grpc

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/status"

	"github.com/jukylin/esim/config"
	logx "github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg/uid"

	"github.com/davecgh/go-spew/spew"

	"google.golang.org/grpc/metadata"

	"github.com/jukylin/esim/core/meta"

	"google.golang.org/grpc"
)

// server
func (gs *Server) metaServerData() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var start = time.Now()
		var cmd = meta.MD{}
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			for k, v := range md {
				if _, o := cmd[k]; !o {
					cmd[k] = v[0]
				}
			}
		}

		ctx = meta.NewContext(ctx, cmd)

		// request logger
		gs.logger.Infoc(ctx, "RequestParams: method[%v], server[%v], params:%+v",
			info.FullMethod, info.Server, spew.Sdump(cmd))

		// monitor
		if gs.monitor {
			ServerGRPCReqQPS.Inc(meta.String(ctx, meta.ServiceName), info.FullMethod, meta.String(ctx, meta.AppID))
		}
		resp, err = handler(ctx, req)

		duration := time.Since(start)
		// response logger
		gs.logger.Infoc(ctx, "ResponseParams: method[%v], server[%v], cost[%v], resp_params:%+v",
			info.FullMethod, info.Server, duration.String(), spew.Sdump(resp))

		// monitor
		if gs.monitor {
			ServerGRPCReqDuration.Observe(float64(duration/time.Millisecond),
				meta.String(ctx, meta.ServiceName), info.FullMethod, meta.String(ctx, meta.AppID))
		}

		// check grpc slow
		grpcSlowTime := gs.conf.GetInt64("grpc_server_slow_time")
		if grpcSlowTime != 0 {
			if time.Now().Sub(start) > time.Duration(grpcSlowTime)*time.Millisecond {
				gs.logger.Warnc(ctx, "Slow server %s", info.FullMethod)
			}
		}

		return resp, err
	}
}

func (gc *ClientOptions) metaClientData() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		cmd, err := setClientMetadata(req)
		if err != nil {
			gst, _ := status.FromError(err)
			return errors.WithMessage(err, gst.Message())
		}

		ctx = metadata.NewOutgoingContext(ctx, cmd)

		// request params
		gc.logger.Infoc(ctx, "SendRequestParams: method[%v], params:%+v", method, spew.Sdump(req))

		if err = invoker(ctx, method, req, reply, cc, opts...); err != nil {
			gst, _ := status.FromError(err)
			err = errors.WithMessage(err, gst.Message())
		}

		return err
	}
}

// client
func setClientMetadata(req interface{}) (metadata.MD, error) {
	var cmd = new(meta.CommonHeader)
	var d = metadata.MD{}
	var marshal = func(v interface{}) []byte {
		b, _ := json.Marshal(v)
		return b
	}

	err := json.Unmarshal(marshal(req), cmd)
	if err != nil {
		logx.Errorf("metadata unmarshal err: %v", err)
		return nil, err
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
	if dstSysId := cmd.Head.SrcSysId; dstSysId == "" {
		d.Set(meta.DstSysId, config.GetString("appname"))
	} else {
		d.Set(meta.DstSysId, dstSysId)
	}

	if traceID := cmd.Head.TraceId; traceID == "" {
		traceID = uid.NewUIDRepo().TraceID()
		d.Set(meta.TraceID, traceID)
	} else {
		d.Set(meta.TraceID, cmd.Head.TraceId)
	}

	return d, err
}
