package container

import (
	"sync"

	"github.com/jukylin/esim/core/metrics"

	"github.com/jukylin/esim/core/tracer"

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

func provideConf() config.Config {
	return confFunc()
}

func providePrometheus() *metrics.Prometheus {
	return metrics.NewPrometheus()
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
