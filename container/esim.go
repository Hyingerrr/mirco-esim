package container

import (
	"net/http"
	"strings"
	"sync"

	config2 "github.com/jukylin/esim/core/config"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/jukylin/esim/metrics"

	"github.com/google/wire"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	eot "github.com/jukylin/esim/opentracing"
	"github.com/opentracing/opentracing-go"
)

var esimOnce sync.Once
var onceEsim *Esim

const defaultAppname = "esim"
const defaultPrometheusHTTPArrd = "9002"

// Esim init start.
type Esim struct {
	prometheus *metrics.Prometheus

	Logger log.Logger

	Conf config2.Config

	Tracer opentracing.Tracer

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

var confFunc = func() config2.Config {
	return config.NewMemConfig()
}

func SetConfFunc(conf func() config2.Config) {
	confFunc = conf
}
func provideConf() config2.Config {
	return confFunc()
}

var prometheusFunc = func(conf config2.Config, logger log.Logger) *metrics.Prometheus {
	addr := conf.GetString("prometheus_http_addr")
	if addr == "" {
		addr = defaultPrometheusHTTPArrd
	}

	if in := strings.Index(addr, ":"); in < 0 {
		addr = ":" + addr
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Panicf(http.ListenAndServe(addr, nil).Error())
	}()

	logger.Info("Prometheus Server Init Success")

	return &metrics.Prometheus{}
}

func providePrometheus(conf config2.Config, logger log.Logger) *metrics.Prometheus {
	return prometheusFunc(conf, logger)
}

var loggerFunc = func(conf config2.Config) log.Logger {
	var loggerOptions log.LoggerOptions
	logger := log.NewLogger(
		loggerOptions.WithLoggerConf(conf),
		loggerOptions.WithDebug(conf.GetBool("debug")),
		//loggerOptions.WithJSON(conf.GetString("runmode") == "pro"),
	)
	return logger
}

func SetLogger(logger func(config2.Config) log.Logger) {
	loggerFunc = logger
}
func provideLogger(conf config2.Config) log.Logger {
	return loggerFunc(conf)
}

var tracerFunc = func(conf config2.Config, logger log.Logger) opentracing.Tracer {
	var appname string
	if conf.GetString("appname") != "" {
		appname = conf.GetString("appname")
	} else {
		appname = defaultAppname
	}

	logger.Infof("appname[%s]", appname)
	return eot.NewTracer(appname, logger)
}

func SetTracer(tracer func(config2.Config, log.Logger) opentracing.Tracer) {
	tracerFunc = tracer
}
func provideTracer(conf config2.Config, logger log.Logger) opentracing.Tracer {
	return tracerFunc(conf, logger)
}

func provideAppName(conf config2.Config) string {
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
