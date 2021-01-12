package grpc

import (
	"net"
	"time"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	opentracing2 "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	server *grpc.Server

	logger log.Logger

	conf config.Config

	interceptors []grpc.UnaryServerInterceptor

	opts []grpc.ServerOption

	target string

	tracer opentracing2.Tracer

	monitor bool
}

type ServerOption func(c *Server)

type ServerOptions struct{}

func NewServer(target string, options ...ServerOption) *Server {
	s := &Server{}

	s.target = target

	for _, option := range options {
		option(s)
	}

	if s.logger == nil {
		s.logger = log.NewLogger()
	}

	if s.conf == nil {
		s.conf = config.NewNullConfig()
	}

	if s.tracer == nil {
		s.tracer = opentracing2.NoopTracer{}
	}

	if s.target == "" {
		s.target = s.conf.GetString("grpc_server_tcp")
	}

	s.monitor = s.conf.GetBool("grpc_server_metrics")

	baseOpts := []grpc.ServerOption{
		grpc.ConnectionTimeout(s.setConnTimeout() * time.Second),
		grpc.KeepaliveParams(s.setKeepAliveParams()),
		grpc.UnaryInterceptor(s.handlerInterceptor),
	}

	if len(s.opts) > 0 {
		baseOpts = append(baseOpts, s.opts...)
	}

	s.server = grpc.NewServer(baseOpts...)

	s.Use(s.metaServerData(), s.tracerID(), s.recovery())

	if s.conf.GetBool("grpc_server_tracer") {
		s.Use(otgrpc.OpenTracingServerInterceptor(s.tracer))
	}

	return s
}

func (ServerOptions) WithServerConf(conf config.Config) ServerOption {
	return func(g *Server) {
		g.conf = conf
	}
}

func (ServerOptions) WithServerLogger(logger log.Logger) ServerOption {
	return func(g *Server) {
		g.logger = logger
	}
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

func (ServerOptions) WithTracer(tracer opentracing2.Tracer) ServerOption {
	return func(g *Server) {
		g.tracer = tracer
	}
}

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

func ServerStubs(stubsFunc func(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error)) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		return stubsFunc(ctx, req, info, handler)
	}
}

func (gs *Server) Start() {
	lis, err := net.Listen("tcp", gs.target)
	if err != nil {
		gs.logger.Panicf("Failed to listen: %s", err.Error())
	}

	// Register reflection service on gRPC server.
	reflection.Register(gs.server)

	gs.logger.Infof("Grpc server starting %s:%s", gs.conf.GetString("appname"), gs.target)
	go func() {
		if err := gs.server.Serve(lis); err != nil {
			gs.logger.Panicf("Failed to start server: %s", err.Error())
		}
	}()
}

func (gs *Server) GracefulShutDown() {
	gs.server.GracefulStop()
}

func (gs *Server) Server() *grpc.Server {
	return gs.server
}

// todo
// 1. 监控
// 2. 超时时间
// 3. metadata
// 4. grpc_client优化
