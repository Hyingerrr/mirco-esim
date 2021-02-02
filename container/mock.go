package container

import (
	"github.com/google/wire"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/core/tracer"
	"github.com/jukylin/esim/metrics"
)

func provideMockConf() config.Config {
	conf := config.NewMemConfig()
	conf.Set("debug", true)
	return conf
}

func provideMockProme(conf config.Config) *metrics.Prometheus {
	return prometheusFunc(conf, nil)
}

func provideMockAppName(conf config.Config) string {
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
