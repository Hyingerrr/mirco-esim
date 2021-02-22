package grpc

import (
	"net"

	"google.golang.org/grpc/keepalive"

	logx "github.com/jukylin/esim/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	server *grpc.Server

	interceptors []grpc.UnaryServerInterceptor

	opts []grpc.ServerOption

	config *ServerConfig
}

type ServerOption func(c *Server)

type ServerOptions struct{}

func NewServer(options ...ServerOption) *Server {
	s := &Server{}

	for _, option := range options {
		option(s)
	}

	// set default config
	s.setServerConfig()

	baseOpts := []grpc.ServerOption{
		grpc.ConnectionTimeout(s.config.DialTimeout),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Timeout: s.config.KeepTimeOut,
			Time:    s.config.KeepTime,
		}),
		grpc.UnaryInterceptor(s.handlerInterceptor),
	}

	if len(s.opts) > 0 {
		baseOpts = append(baseOpts, s.opts...)
	}

	s.server = grpc.NewServer(baseOpts...)

	s.Use(recoverServerInterceptor(), tracerIDServerInterceptor())
	// timeout
	s.Use(timeoutUnaryServerInterceptor(s.config.Timeout))

	if s.config.Debug {
		s.Use(debugUnaryServerInterceptor(s.config.SlowTime))
	}

	if s.config.Validate {
		s.Use(validateServerInterceptor())
	}

	if s.config.Tracer {
		s.Use(traceUnaryServerInterceptor)
	}

	if s.config.Metrics {
		s.Use(metricUnaryServerInterceptor)
	}

	return s
}

func (ServerOptions) WithUnarySrvItcp(options ...grpc.UnaryServerInterceptor) ServerOption {
	return func(g *Server) {
		g.interceptors = options
	}
}

func (ServerOptions) WithServerOption(options ...grpc.ServerOption) ServerOption {
	return func(g *Server) {
		g.opts = options
	}
}

// handler chain
func (gs *Server) handlerInterceptor(ctx context.Context, req interface{}, args *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var (
		i     int
		chain grpc.UnaryHandler
	)

	n := len(gs.interceptors)
	if n == 0 {
		return handler(ctx, req)
	}

	chain = func(ic context.Context, ir interface{}) (interface{}, error) {
		if i == n-1 {
			return handler(ic, ir)
		}
		i++
		return gs.interceptors[i](ic, ir, args, chain)
	}

	return gs.interceptors[0](ctx, req, args, chain)
}

func (gs *Server) Use(interceptors ...grpc.UnaryServerInterceptor) *Server {
	finalSize := len(gs.interceptors) + len(interceptors)
	if finalSize >= _abortIndex {
		panic("ESIM: server use too many interceptors")
	}

	mergedHandlers := make([]grpc.UnaryServerInterceptor, finalSize)
	copy(mergedHandlers, gs.interceptors)
	copy(mergedHandlers[len(gs.interceptors):], interceptors)
	gs.interceptors = mergedHandlers
	return gs
}

func (gs *Server) Start() {
	lis, err := net.Listen("tcp", gs.config.Addr)
	if err != nil {
		logx.Panicf("Failed to listen: %s", err.Error())
	}

	// Register reflection service on gRPC server.
	reflection.Register(gs.server)

	logx.Infof("Grpc server starting %s:%s", gs.config.AppName, gs.config.Addr)
	go func() {
		if err := gs.server.Serve(lis); err != nil {
			logx.Panicf("Failed to start server: %s", err.Error())
		}
	}()
}

func (gs *Server) GracefulShutDown() {
	gs.server.GracefulStop()
}

func (gs *Server) Server() *grpc.Server {
	return gs.server
}
