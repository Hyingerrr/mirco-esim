package rest

import (
	config2 "github.com/jukylin/esim/core/config"
)

type (
	Server struct {
		engine *engine
	}

	RunOptions func(*Server)
)

func NewServer(opts ...RunOptions) *Server {
	server := &Server{
		engine: newEngine(),
	}

	for _, opt := range opts {
		opt(server)
	}

	server.engine.Start()
	return server
}

func (RunOptions) ServerWithConf(conf config2.Config) {

}
