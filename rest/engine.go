package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/core/xenv"
	logx "github.com/jukylin/esim/log"
	"github.com/jukylin/esim/rest/handler"
)

type engine struct {
	gin *gin.Engine
}

type Option func(*engine)

func NewGinEngine(opts ...Option) *engine {
	e := &engine{gin: gin.Default()}

	for _, opt := range opts {
		opt(e)
	}

	m := xenv.SetRunMode(config.GetString("runmode"))
	if !m.IsValid() {
		logx.Panicf("RunMode InValid")
	}

	if xenv.IsPro() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	e.bindMiddleware()

	// other something todo
	return e
}

func (en *engine) bindMiddleware() {
	en.gin.Use(handler.TracerID(), handler.Recover())

	if config.GetBool("http_metrics") {
		en.gin.Use(handler.SetMetadata(), handler.HttpMonitor())
	}

	if config.GetBool("http_tracer") {
		en.gin.Use(handler.HttpTracer())
	}
}

func (en *engine) Gine() *gin.Engine {
	return en.gin
}

func (en *engine) Use(middleware ...gin.HandlerFunc) {
	en.gin.Use(middleware...)
}
