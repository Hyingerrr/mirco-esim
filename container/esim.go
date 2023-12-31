package container

import (
	"sync"

	"github.com/Hyingerrr/mirco-esim/core/metrics"

	"github.com/Hyingerrr/mirco-esim/core/tracer"

	"github.com/Hyingerrr/mirco-esim/config"
	"github.com/Hyingerrr/mirco-esim/log"

	"github.com/google/wire"
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
