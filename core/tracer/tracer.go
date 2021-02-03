package tracer

import (
	"io"

	"github.com/uber/jaeger-client-go"
	jconfig "github.com/uber/jaeger-client-go/config"

	"github.com/opentracing/opentracing-go"
)

type EsimTracer struct {
	opentracing.Tracer
	io.Closer
}

func InitTracer() *EsimTracer {
	config := initDefaultConfig()
	config.WithOption(jconfig.Logger(jaeger.StdLogger))
	tracer, closer := config.Build()
	// set global
	opentracing.SetGlobalTracer(tracer)
	return &EsimTracer{Tracer: tracer, Closer: closer}
}

func HeaderExtractor(hd map[string][]string) opentracing.StartSpanOption {
	spCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders,
		MetadataReaderWriter{MD: hd})
	if err != nil {
		return NullStartSpanOption{}
	}

	return opentracing.ChildOf(spCtx)
}

// NullStartSpanOption ...
type NullStartSpanOption struct{}

// Apply ...
func (sso NullStartSpanOption) Apply(options *opentracing.StartSpanOptions) {}
