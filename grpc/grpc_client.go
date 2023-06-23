package grpc

import (
	"time"

	"google.golang.org/grpc/keepalive"

	logx "github.com/Hyingerrr/mirco-esim/log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type Client struct {
	conn       *grpc.ClientConn
	clientOpts *ClientOptions
}

type ClientOptions struct {
	opts   []grpc.DialOption
	config *ClientConfig
}

type ClientOptional func(c *ClientOptions)

func NewClientOptions(options ...ClientOptional) *ClientOptions {
	c := &ClientOptions{}

	for _, option := range options {
		option(c)
	}

	c.setClientConfig()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                c.config.KeepTime,
			Timeout:             c.config.KeepTimeOut,
			PermitWithoutStream: c.config.PermitWithoutStream,
		}),
		grpc.WithChainUnaryInterceptor(
			timeOutUnaryClientInterceptor(c.config.Timeout), metadataHandler()),
	}

	if c.config.Debug {
		opts = append(opts,
			grpc.WithChainUnaryInterceptor(debugUnaryClientInterceptor(c.config.SlowTime)))
	}

	if c.config.Tracer {
		opts = append(opts,
			grpc.WithChainUnaryInterceptor(traceUnaryClientInterceptor()))
	}

	if c.config.Metrics {
		opts = append(opts,
			grpc.WithChainUnaryInterceptor(metricUnaryClientInterceptor()))
	}

	c.opts = append(c.opts, opts...)

	return c
}

func WithDialOptions(options ...grpc.DialOption) ClientOptional {
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
	var err error

	// connect timeout ctrl
	if dt := gc.clientOpts.config.DialTimeout; dt > 0 {
		ctx, cancel = context.WithTimeout(ctx, dt)
		defer cancel()

		// grpc.WithBlock()等待链接建立完成; 否则dialTimeout无效
		gc.clientOpts.opts = append(gc.clientOpts.opts, grpc.WithBlock())

		//todo debug
		dl, _ := ctx.Deadline()
		logx.Infoc(ctx, "拨号deadline:%v", dl.String())
	}

	gc.conn, err = grpc.DialContext(ctx, target, gc.clientOpts.opts...)
	if err != nil {
		logx.Errorc(ctx, "grpc dial error: %v", err)
		return nil
	}

	return gc.conn
}

func (gc *Client) Close() {
	_ = gc.conn.Close()
}

type TimeoutCallOption struct {
	*grpc.EmptyCallOption
	Timeout time.Duration // ms
}

func WithTimeout(timeout time.Duration) *TimeoutCallOption {
	return &TimeoutCallOption{
		EmptyCallOption: &grpc.EmptyCallOption{},
		Timeout:         timeout,
	}
}
