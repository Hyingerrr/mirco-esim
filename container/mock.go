package container

import (
	"github.com/google/wire"
	"github.com/jukylin/esim/config"
	config2 "github.com/jukylin/esim/core/config"
	"github.com/jukylin/esim/metrics"
	"github.com/opentracing/opentracing-go"
)

func provideMockConf() config2.Config {
	conf := config.NewMemConfig()
	conf.Set("debug", true)
	return conf
}

func provideMockProme(conf config2.Config) *metrics.Prometheus {
	return prometheusFunc(conf, nil)
}

func provideMockAppName(conf config2.Config) string {
	return "mocktest"
}

func provideNoopTracer() opentracing.Tracer {
	return opentracing.NoopTracer{}
}

var MockSet = wire.NewSet(
	wire.Struct(new(Esim), "*"),
	provideMockConf,
	provideLogger,
	provideMockProme,
	provideNoopTracer,
	provideMockAppName,
)
