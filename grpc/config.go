package grpc

import (
	"time"

	"google.golang.org/grpc/keepalive"
)

var (
	_abortIndex = 64
)

func (gs *Server) setKeepAliveParams() keepalive.ServerParameters {
	kpTime := gs.conf.GetInt("grpc_server_kp_time")
	if kpTime == 0 {
		kpTime = 60
	}

	kpTimeout := gs.conf.GetInt("grpc_server_kp_time_out")
	if kpTimeout == 0 {
		kpTimeout = 5
	}

	connTimeout := gs.conf.GetInt("grpc_server_conn_time_out")
	if connTimeout == 0 {
		connTimeout = 3
	}

	return keepalive.ServerParameters{
		Time:    time.Duration(kpTime) * time.Second,
		Timeout: time.Duration(kpTimeout) * time.Second,
	}
}

func (gs *Server) setConnTimeout() time.Duration {
	connTimeout := gs.conf.GetInt("grpc_server_conn_time_out")
	if connTimeout == 0 {
		connTimeout = 3
	}

	return time.Duration(connTimeout)
}
