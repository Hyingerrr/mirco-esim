package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/core/xenv"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/rest/handler"
)

type engine struct {
	gin    *gin.Engine
	conf   config.Config
	logger log.Logger
}

type Option func(*engine)

func NewGinEngine(opts ...Option) *engine {
	e := &engine{gin: gin.Default()}

	for _, opt := range opts {
		opt(e)
	}

	m := xenv.SetRunMode(e.conf.GetString("runmode"))
	if !m.IsValid() {
		e.logger.Panicf("RunMode InValid")
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

func WithLogger(logger log.Logger) Option {
	return func(e *engine) {
		e.logger = logger
	}
}

func WithConf(conf config.Config) Option {
	return func(e *engine) {
		e.conf = conf
	}
}

func (en *engine) bindMiddleware() {
	if en.conf.GetBool("http_metrics") {
		en.gin.Use(handler.HttpMonitor())
	}

	if en.conf.GetBool("http_tracer") {
		en.gin.Use(handler.HttpTracer())
	}

	en.gin.Use(handler.TracerID(), handler.Recover(en.logger))

	en.gin.Use(handler.MetaDataCtx())
}

func (en *engine) Gine() *gin.Engine {
	return en.gin
}
