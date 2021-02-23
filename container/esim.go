package container

import (
	"sync"

	"github.com/jukylin/esim/core/tracer"

	"github.com/jukylin/esim/core/metrics"

	"github.com/google/wire"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
)

var (
	esimOnce sync.Once
	onceEsim *Esim
)

// Esim init start.
type Esim struct {
	prometheus *metrics.Prometheus

	Logger log.Logger

	Conf config.Config

	Tracer *tracer.EsimTracer

	AppName string
}

//nolint:varcheck,unused,deadcode
var esimSet = wire.NewSet(
	wire.Struct(new(Esim), "*"),
	provideConf,
	provideLogger,
	providePrometheus,
	provideTracer,
	provideAppName,
)

var confFunc = func() config.Config {
	return config.NewMemConfig()
}

func SetConfFunc(conf func() config.Config) {
	confFunc = conf
}
func provideConf() config.Config {
	return confFunc()
}

func providePrometheus() *metrics.Prometheus {
	return metrics.NewPrometheus()
}

var loggerFunc = func(conf config.Config) log.Logger {
	var loggerOptions log.LoggerOptions
	logger := log.NewLogger(
		loggerOptions.WithLoggerConf(conf),
		loggerOptions.WithDebug(conf.GetBool("debug")),
	)
	return logger
}

func SetLogger(logger func(config.Config) log.Logger) {
	loggerFunc = logger
}
func provideLogger(conf config.Config) log.Logger {
	return loggerFunc(conf)
}

func provideTracer() *tracer.EsimTracer {
	return tracer.InitTracer()
}

func provideAppName(conf config.Config) string {
	return conf.GetString("appname")
}

// Esim init end.
func NewEsim() *Esim {
	esimOnce.Do(func() {
		onceEsim = initEsim()
	})

	return onceEsim
}

func (e *Esim) String() string {
	return "Esim 基础框架;"
}

func AppName() string {
	return onceEsim.AppName
}
