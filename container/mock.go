package container

import (
	"github.com/Hyingerrr/mirco-esim/config"
	"github.com/Hyingerrr/mirco-esim/core/metrics"
	"github.com/Hyingerrr/mirco-esim/core/tracer"

	"github.com/google/wire"
)

func provideMockConf() config.Config {
	conf := config.NewMemConfig()
	conf.Set("debug", true)
	return conf
}

func provideMockProme() *metrics.Prometheus {
	return metrics.NewPrometheus()
}

func provideMockAppName() string {
	return "mocktest"
}

func provideNoopTracer() *tracer.EsimTracer {
	return &tracer.EsimTracer{}
}

var MockSet = wire.NewSet(
	wire.Struct(new(Esim), "*"),
	provideMockConf,
	provideLogger,
	provideMockProme,
	provideNoopTracer,
	provideMockAppName,
)
