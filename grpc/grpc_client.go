package grpc

import (
	"time"

	"github.com/davecgh/go-spew/spew"
	ggp "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/opentracing"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type Client struct {
	conn *grpc.ClientConn

	cancel context.CancelFunc

	logger log.Logger

	clientOpts *ClientOptions
}

type ClientOptions struct {
	clientMetrics *ggp.ClientMetrics

	tracer opentracing2.Tracer

	logger log.Logger

	conf config.Config

	opts []grpc.DialOption
}

type ClientOptional func(c *ClientOptions)

type ClientOptionals struct{}

func NewClientOptions(options ...ClientOptional) *ClientOptions {
	co := &ClientOptions{}

	for _, option := range options {
		option(co)
	}

	if co.logger == nil {
		co.logger = log.NewLogger()
	}

	if co.conf == nil {
		co.conf = config.NewMemConfig()
	}

	if co.tracer == nil {
		co.tracer = opentracing.NewTracer("grpc_client", co.logger)
	}

	if co.clientMetrics == nil {
		ggp.EnableClientHandlingTimeHistogram(ggp.WithHistogramBuckets(prometheus.DefBuckets))
		co.clientMetrics = ggp.DefaultClientMetrics
	}

	keepAliveClient := keepalive.ClientParameters{}
	ClientKpTime := co.conf.GetInt("grpc_client_kp_time")
	if ClientKpTime == 0 {
		ClientKpTime = 60
	}
	keepAliveClient.Time = time.Duration(ClientKpTime) * time.Second

	ClientKpTimeOut := co.conf.GetInt("grpc_client_kp_time_out")
	if ClientKpTimeOut == 0 {
		ClientKpTimeOut = 5
	}
	keepAliveClient.Timeout = time.Duration(ClientKpTimeOut) * time.Second

	ClientPermitWithoutStream := co.conf.GetBool("grpc_client_permit_without_stream")
	keepAliveClient.PermitWithoutStream = ClientPermitWithoutStream

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepAliveClient),
	}

	opts = append(opts, grpc.WithChainUnaryInterceptor(co.metaClientData()))

	if co.conf.GetBool("grpc_client_tracer") {
		tracerInterceptor := otgrpc.OpenTracingClientInterceptor(co.tracer)
		opts = append(opts, grpc.WithChainUnaryInterceptor(tracerInterceptor))
	}

	if co.conf.GetBool("grpc_client_metrics") {
		opts = append(opts, grpc.WithChainUnaryInterceptor(
			co.clientMetrics.UnaryClientInterceptor()))
	}

	if co.conf.GetBool("grpc_client_check_slow") {
		opts = append(opts, grpc.WithChainUnaryInterceptor(co.checkClientSlow()))
	}

	if co.conf.GetBool("grpc_client_debug") {
		opts = append(opts, grpc.WithChainUnaryInterceptor(co.clientDebug()))
	}

	co.opts = append(opts, co.opts...)

	return co
}

func (ClientOptionals) WithConf(conf config.Config) ClientOptional {
	return func(g *ClientOptions) {
		g.conf = conf
	}
}

func (ClientOptionals) WithLogger(logger log.Logger) ClientOptional {
	return func(g *ClientOptions) {
		g.logger = logger
	}
}

func (ClientOptionals) WithTracer(tracer opentracing2.Tracer) ClientOptional {
	return func(g *ClientOptions) {
		g.tracer = tracer
	}
}

func (ClientOptionals) WithMetrics(metrics *ggp.ClientMetrics) ClientOptional {
	return func(g *ClientOptions) {
		g.clientMetrics = metrics
	}
}

func (ClientOptionals) WithDialOptions(options ...grpc.DialOption) ClientOptional {
	return func(g *ClientOptions) {
		g.opts = options
	}
}

// NewClient create Client for business.
// clientOptions clientOptions can not nil.
func NewClient(clientOptions *ClientOptions) *Client {
	c := &Client{}

	c.clientOpts = clientOptions

	return c
}

func (gc *Client) DialContext(ctx context.Context, target string) *grpc.ClientConn {
	var cancel context.CancelFunc

	ClientConnTimeOut := gc.clientOpts.conf.GetInt("grpc_client_conn_time_out")
	if ClientConnTimeOut == 0 {
		ClientConnTimeOut = 3
		ctx, cancel = context.WithTimeout(ctx, time.Duration(ClientConnTimeOut)*time.Second)
		gc.cancel = cancel
	}

	conn, err := grpc.DialContext(ctx, target, gc.clientOpts.opts...)
	if err != nil {
		gc.logger.Errorf(err.Error())
		return nil
	}
	gc.conn = conn

	return conn
}

func (gc *Client) Close() {
	gc.conn.Close()
	gc.cancel()
}

func (gc *ClientOptions) checkClientSlow() func(ctx context.Context,
	method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ClientSlowTime := gc.conf.GetInt64("grpc_client_slow_time")

		beginTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		endTime := time.Now()

		if ClientSlowTime != 0 {
			if endTime.Sub(beginTime) > time.Duration(ClientSlowTime)*time.Millisecond {
				gc.logger.Warnc(ctx, "slow client grpc_handle %s", method)
			}
		}

		return err
	}
}

func (gc *ClientOptions) clientDebug() func(ctx context.Context,
	method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		beginTime := time.Now()
		gc.logger.Debugc(ctx, "Grpc client start %s : %s, req : %s",
			cc.Target(), method, spew.Sdump(req))

		err := invoker(ctx, method, req, reply, cc, opts...)

		endTime := time.Now()
		gc.logger.Debugc(ctx, "Grpc client end [%v] %s : %s, reply : %s",
			endTime.Sub(beginTime).String(), cc.Target(), method, spew.Sdump(reply))

		return err
	}
}

func ClientStubs(stubsFunc func(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error) func(ctx context.Context,
	method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{},
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := stubsFunc(ctx, method, req, reply, cc, invoker, opts...)
		return err
	}
}

func slowRequest(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	time.Sleep(20 * time.Millisecond)
	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}

//nolint:deadcode,unused
func timeoutRequest(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	time.Sleep(10 * time.Second)
	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}
