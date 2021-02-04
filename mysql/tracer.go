package mysql

import (
	"context"

	"github.com/jinzhu/gorm"
	"github.com/opentracing/opentracing-go"
)

const (
	OpenTracingContextKey     = "gorm:opentracing_context"
	OpenTracingSpanContextKey = "gorm:opentracing_spanner"
)

type TraceClient struct {
	*gorm.DB
	trace *TraceContext
}

type TraceContext struct {
	tracer  opentracing.Tracer
	spanCtx opentracing.SpanContext
}

func (ctx *TraceContext) StartSpan(name string, tags ...opentracing.StartSpanOption) opentracing.Span {
	// create root span
	if ctx.spanCtx == nil {
		return ctx.tracer.StartSpan(name)
	}

	return ctx.tracer.StartSpan(name, tags...)
}

func (c *Client) Trace(ctx context.Context, db *gorm.DB) *TraceClient {
	if ctx != nil {
		span := opentracing.SpanFromContext(ctx)
		if span != nil {
			return c.TraceWithSpanContext(span.Context(), db)
		}
	}

	return c.TraceWithSpanContext(nil, db)
}

func (c *Client) TraceWithSpanContext(ctx opentracing.SpanContext, db *gorm.DB) *TraceClient {
	c.RegisterTraceCallbacks(db)

	trace := TraceContext{
		tracer:  opentracing.GlobalTracer(),
		spanCtx: ctx,
	}

	return &TraceClient{
		DB:    db.Set(OpenTracingContextKey, trace),
		trace: &trace,
	}
}
