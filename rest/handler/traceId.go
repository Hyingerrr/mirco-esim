package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/opentracing/opentracing-go/ext"

	"github.com/gin-gonic/gin"
	tracerid "github.com/jukylin/esim/pkg/tracer-id"
	"github.com/opentracing/opentracing-go"
)

func TracerID() gin.HandlerFunc {
	tracerID := tracerid.TracerID()
	return func(c *gin.Context) {
		sp := opentracing.SpanFromContext(c.Request.Context())
		if sp == nil {
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(),
				tracerid.ActiveEsimKey, tracerID()))
		}

		c.Next()
	}
}

func HttpTracer() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tracer = opentracing.GlobalTracer()
		var span opentracing.Span
		spCtx, err := tracer.Extract(opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header))
		if err != nil {
			// global start
			span = tracer.StartSpan("HTTP " + c.Request.URL.Path)
			defer span.Finish()
		} else {
			// open tracing start
			span = opentracing.StartSpan(
				"HTTP "+c.Request.URL.Path,
				opentracing.ChildOf(spCtx),
			)
			defer span.Finish()
		}

		ext.HTTPMethod.Set(span, c.Request.Method)
		ext.HTTPUrl.Set(span, c.Request.URL.Path)
		ext.Component.Set(span, "http")
		ext.SpanKind.Set(span, "server")
		ip, _ := strconv.ParseUint(c.ClientIP(), 10, 64)
		ext.PeerHostIPv4.Set(span, uint32(ip))

		c.Request = c.Request.WithContext(opentracing.ContextWithSpan(c.Request.Context(), span))
		c.Next()

		ext.HTTPStatusCode.Set(span, uint16(c.Writer.Status()))
		if c.Writer.Status() >= http.StatusInternalServerError {
			ext.Error.Set(span, true)
		}
		span.Finish()
	}
}
