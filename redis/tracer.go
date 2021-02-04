package redis

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
)

type TraceClient struct {
	*Client
	tracer  opentracing.Tracer
	spanCtx opentracing.SpanContext
}

func (c *Client) withTrace(ctx context.Context) *TraceClient {
	if ctx != nil {
		span := opentracing.SpanFromContext(ctx)
		if span != nil {
			return c.TraceWithSpanContext(span.Context())
		}
	}

	return c.TraceWithSpanContext(nil)
}

func (c *Client) TraceWithSpanContext(spCtx opentracing.SpanContext) *TraceClient {
	return &TraceClient{
		Client:  c,
		tracer:  opentracing.GlobalTracer(),
		spanCtx: spCtx,
	}
}

func getStatement(command string, args ...interface{}) string {
	res := command
	if len(args) == 1 || len(args) > 3 {
		res = fmt.Sprintf("%s %v", command, args[0])
	}

	if len(args) == 2 {
		res = fmt.Sprintf("%s %v %v", command, args[0], args[1])
	}

	if len(args) == 3 {
		res = fmt.Sprintf("%s %v %v %v", command, args[0], args[1], args[2])
	}

	return res
}
