package tracer

import (
	"io"
	"time"

	"github.com/jukylin/esim/pkg/hepler"

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

func CustomTag(key string, val interface{}) opentracing.Tag {
	return opentracing.Tag{
		Key:   key,
		Value: val,
	}
}

func TagDbSchema(schema string) opentracing.Tag {
	return opentracing.Tag{
		Key:   "db.schema",
		Value: schema,
	}
}

func TagComponent(val string) opentracing.Tag {
	return opentracing.Tag{
		Key:   "component",
		Value: val,
	}
}

func TagLocalIPV4() opentracing.Tag {
	ip, _ := hepler.GetLocalIp()
	return opentracing.Tag{
		Key:   "peer.ipv4",
		Value: ip,
	}
}

func TagStartTime() opentracing.StartSpanOption {
	return opentracing.StartTime(time.Now())
}

func TagFinishTime() opentracing.FinishOptions {
	return opentracing.FinishOptions{FinishTime: time.Time{}}
}
