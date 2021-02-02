package tracer

import (
	"io"
	"os"

	"github.com/opentracing/opentracing-go"
)

type EsimTracer struct {
	opentracing.Tracer
	io.Closer
}

func InitTracer() *EsimTracer {
	_ = os.Setenv("APP_NAME", "TestEsim")
	config := initDefaultConfig()
	tracer, closer := config.Build()
	opentracing.SetGlobalTracer(tracer)
	return &EsimTracer{
		Tracer: tracer,
		Closer: closer,
	}
}

// NullStartSpanOption ...
type NullStartSpanOption struct{}

// Apply ...
func (sso NullStartSpanOption) Apply(options *opentracing.StartSpanOptions) {}
