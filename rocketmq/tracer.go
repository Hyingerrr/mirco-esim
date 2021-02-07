package rocketmq

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

// Publisher
type TracePublisher struct {
	*Publisher
	tracer  opentracing.Tracer
	spanCtx opentracing.SpanContext
}

func (p *Publisher) withTrace(ctx context.Context) *TracePublisher {
	if ctx != nil {
		span := opentracing.SpanFromContext(ctx)
		if span != nil {
			return p.TraceWithSpanContext(span.Context())
		}
	}

	return p.TraceWithSpanContext(nil)
}

func (p *Publisher) TraceWithSpanContext(spCtx opentracing.SpanContext) *TracePublisher {
	return &TracePublisher{
		Publisher: p,
		tracer:    opentracing.GlobalTracer(),
		spanCtx:   spCtx,
	}
}

// Subscriber
type TraceSubscriber struct {
	*SubscribeEngine
	tracer opentracing.Tracer
}

func (se *SubscribeEngine) withTraceRootSpan() opentracing.Span {
	tc := &TraceSubscriber{
		SubscribeEngine: se,
		tracer:          opentracing.GlobalTracer(),
	}

	return tc.tracer.StartSpan("RocketMQ_Subscriber")
}
