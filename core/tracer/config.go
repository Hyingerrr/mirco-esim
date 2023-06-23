package tracer

import (
	"io"
	"os"
	"time"

	logx "github.com/Hyingerrr/mirco-esim/log"

	"github.com/Hyingerrr/mirco-esim/config"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jconfig "github.com/uber/jaeger-client-go/config"
)

type Config struct {
	ServiceName string
	RPCMetrics  bool
	Sampler     *jconfig.SamplerConfig
	Reporter    *jconfig.ReporterConfig
	Headers     *jaeger.HeadersConfig
	tags        []opentracing.Tag
	options     []jconfig.Option
}

func initDefaultConfig() *Config {
	hn, _ := os.Hostname()
	return &Config{
		ServiceName: config.GetString("appname"),
		RPCMetrics:  config.GetBool("rpc_metrics"),
		tags: []opentracing.Tag{
			{Key: "hostname", Value: hn},
		},
		// const 固定采样 Param 1全采样 0不采样
		// probabilistic 按百分比  Param 0.1则随机采十分之一的样本
		// ratelimiting 采样速度限制 Param 2.0 每秒采样两个traces
		// remote 动态获取采样率 默认配置，可以通过配置从 Agent 中获取采样率的动态设置
		Sampler: &jconfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jconfig.ReporterConfig{
			BufferFlushInterval: time.Second,
			LogSpans:            false,
			LocalAgentHostPort:  config.GetString("tracer_jaeger_upd"),
		},
		Headers: &jaeger.HeadersConfig{
			TraceContextHeaderName:   "x-trace-id",
			TraceBaggageHeaderPrefix: "ctx-",
		},
	}
}

func (config *Config) WithTag(tags ...opentracing.Tag) *Config {
	if config.tags == nil {
		config.tags = make([]opentracing.Tag, 0)
	}

	config.tags = append(config.tags, tags...)
	return config
}

func (config *Config) WithOption(opts ...jconfig.Option) *Config {
	if config.options == nil {
		config.options = make([]jconfig.Option, 0)
	}
	config.options = append(config.options, opts...)
	return config
}

func (config *Config) Build() (opentracing.Tracer, io.Closer) {
	var configuration = jconfig.Configuration{
		ServiceName: config.ServiceName,
		Disabled:    false,
		RPCMetrics:  config.RPCMetrics,
		Tags:        config.tags,
		Sampler:     config.Sampler,
		Reporter:    config.Reporter,
		Headers:     config.Headers,
	}

	tracer, closer, err := configuration.NewTracer(config.options...)
	if err != nil {
		logx.Panicf("new jaeger panic: %v", err)
	}

	opentracing.SetGlobalTracer(tracer)

	return tracer, closer
}
