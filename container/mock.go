package container

import (
	"github.com/google/wire"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/core/metrics"
	"github.com/jukylin/esim/core/tracer"
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
