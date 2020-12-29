package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/jukylin/esim/core/config"
	"github.com/jukylin/esim/core/rest/handler"
)

type engine struct {
	conf   config.Config
	routes []Route
	ngin   *gin.Engine
}

func newEngine() *engine {
	return &engine{
		ngin: gin.Default(),
		conf: config.LoadConf(),
	}
}

func (en *engine) Start() {
	// bind middleware
	en.bindMiddleware()
}

func (en *engine) bindMiddleware() {
	if en.conf.GetBool("http_metrics") {
		en.ngin.Use(handler.HttpMonitor())
	}

	if en.conf.GetBool("http_tracer") {
		en.ngin.Use(handler.HttpTracer())
	}

	en.ngin.Use(handler.TracerID())
}
