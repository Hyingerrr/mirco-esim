package grpc

import (
	"time"

	"google.golang.org/grpc/keepalive"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	logx "github.com/jukylin/esim/log"
	"github.com/jukylin/esim/opentracing"
	opentracing2 "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Client struct {
	conn *grpc.ClientConn

	clientOpts *ClientOptions
}

type ClientOptions struct {
	tracer opentracing2.Tracer

	opts []grpc.DialOption

	config *ClientConfig
}

type ClientOptional func(c *ClientOptions)

type ClientOptionals struct{}

func NewClientOptions(options ...ClientOptional) *ClientOptions {
	c := &ClientOptions{}

	for _, option := range options {
		option(c)
	}

	if c.tracer == nil {
		c.tracer = opentracing.NewTracer("grpc_client", logx.NewLogger())
	}

	c.setClientConfig()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                c.config.KeepTime,
			Timeout:             c.config.KeepTimeOut,
			PermitWithoutStream: c.config.PermitWithoutStream,
		}),
		grpc.WithChainUnaryInterceptor(c.addClientDebug(), c.handleClient()),
	}

	// todo
	if c.config.Tracer {
		tracerInterceptor := otgrpc.OpenTracingClientInterceptor(c.tracer)
		opts = append(opts, grpc.WithChainUnaryInterceptor(tracerInterceptor))
	}

	c.opts = append(c.opts, opts...)

	return c
}

func (ClientOptionals) WithTracer(tracer opentracing2.Tracer) ClientOptional {
	return func(g *ClientOptions) {
		g.tracer = tracer
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
	// connect timeout ctrl
	ctx, cancel = context.WithTimeout(ctx, gc.clientOpts.config.DialTimeout)
	defer cancel()

	//todo debug
	dl, _ := ctx.Deadline()
	logx.Infoc(ctx, "拨号deadline:%v", dl.String())

	conn, err := grpc.DialContext(ctx, target, gc.clientOpts.opts...)
	if err != nil {
		logx.Errorc(ctx, "grpc dial error: %v", err)
		return nil
	}
	gc.conn = conn

	return conn
}

func (gc *Client) Close() {
	gc.conn.Close()
}

type TimeoutCallOption struct {
	*grpc.EmptyCallOption
	Timeout time.Duration
}

func WithTimeout(timeout time.Duration) *TimeoutCallOption {
	return &TimeoutCallOption{
		EmptyCallOption: &grpc.EmptyCallOption{},
		Timeout:         timeout,
	}
}
