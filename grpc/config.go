package grpc

import (
	"strings"
	"time"

	logx "github.com/jukylin/esim/log"

	"github.com/jukylin/esim/config"
)

var (
	_abortIndex = 64
)

type ServerConfig struct {
	AppName string
	Addr    string
	Port    string
	// handle
	Timeout time.Duration // grpc_server_timeout
	// conn
	DialTimeout time.Duration // ms
	// keepalive
	KeepTime    time.Duration // s
	KeepTimeOut time.Duration // s
	// debug
	Debug bool
	// metric
	Metrics bool
	// check slow; ms
	SlowTime time.Duration
	// tracer
	Tracer bool
	// validate
	Validate bool
}

func (gs *Server) setServerConfig() {
	s := &ServerConfig{}
	s.Debug = config.GetBool("grpc_server_debug")
	s.Metrics = config.GetBool("grpc_server_metrics")
	s.Tracer = config.GetBool("grpc_server_tracer")
	s.Validate = config.GetBool("grpc_server_validate")
	s.AppName = config.GetString("appname")
	s.Addr = config.GetString("grpc_server_tcp")
	if s.Addr == "" {
		logx.Panicf("grpc addr is empty")
	}
	if in := strings.Index(s.Addr, ":"); in < 0 {
		s.Addr = ":" + s.Addr
	}

	s.Timeout = config.GetDuration("grpc_server_timeout") * time.Millisecond
	if s.Timeout == 0 {
		s.Timeout = 1000 * time.Millisecond
	}

	s.DialTimeout = config.GetDuration("grpc_server_conn_time_out") * time.Millisecond
	if s.DialTimeout == 0 {
		s.DialTimeout = 2000 * time.Millisecond
	}

	s.KeepTime = config.GetDuration("grpc_server_kp_time") * time.Second
	if s.KeepTime == 0 {
		s.KeepTime = 2 * time.Hour
	}

	s.KeepTimeOut = config.GetDuration("grpc_server_kp_time_out") * time.Second
	if s.KeepTimeOut == 0 {
		s.KeepTimeOut = 20 * time.Second
	}

	s.SlowTime = config.GetDuration("grpc_server_slow_time") * time.Millisecond

	gs.config = s
}

type ClientConfig struct {
	// handle
	Timeout time.Duration
	// conn
	DialTimeout time.Duration // ms
	// keepalive
	KeepTime            time.Duration // s
	KeepTimeOut         time.Duration // s
	PermitWithoutStream bool
	// debug
	Debug bool
	// metric
	Metrics bool
	// check slow; ms
	SlowTime time.Duration
	// tracer
	Tracer bool
}

func (gc *ClientOptions) setClientConfig() {
	s := &ClientConfig{}
	s.Debug = config.GetBool("grpc_client_debug")
	s.Metrics = config.GetBool("grpc_client_metrics")
	s.Tracer = config.GetBool("grpc_client_tracer")
	s.PermitWithoutStream = config.GetBool("grpc_client_permit_without_stream")

	s.Timeout = config.GetDuration("grpc_client_timeout") * time.Millisecond
	if s.Timeout == 0 {
		s.Timeout = 1000 * time.Millisecond
	}

	s.DialTimeout = config.GetDuration("grpc_client_conn_time_out") * time.Millisecond
	if s.DialTimeout == 0 {
		s.DialTimeout = 200 * time.Millisecond
	}

	s.KeepTime = config.GetDuration("grpc_client_kp_time") * time.Second
	if s.KeepTime == 0 {
		s.KeepTime = 3600 * time.Second
	}

	s.KeepTimeOut = config.GetDuration("grpc_client_kp_time_out") * time.Second
	if s.KeepTimeOut == 0 {
		s.KeepTimeOut = 20 * time.Second
	}

	s.SlowTime = config.GetDuration("grpc_client_slow_time") * time.Millisecond

	gc.config = s
}
