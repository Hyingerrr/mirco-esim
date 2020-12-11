package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jukylin/esim/pkg/common/rctx"

	"github.com/gin-gonic/gin"
	tracerid "github.com/jukylin/esim/pkg/tracer-id"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func GinMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			start  = time.Now()
			ctx    = c.Request.Context()
			labels []string
		)

		labels = append(labels,
			c.Request.Host,
			rctx.String(ctx, rctx.LabelTranCd),
			rctx.String(ctx, rctx.LabelProdCd),
			rctx.String(ctx, rctx.LabelAppID),
		)

		c.Next()

		serverReqQPS.Inc(labels...)
		serverReqDuration.Observe(time.Since(start).Seconds(), labels...)
	}
}

func GinTracer(tracer opentracing2.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		spContext, _ := tracer.Extract(opentracing2.HTTPHeaders,
			opentracing2.HTTPHeadersCarrier(c.Request.Header))

		sp := tracer.StartSpan("HTTP "+c.Request.Method,
			ext.RPCServerOption(spContext))

		ext.HTTPMethod.Set(sp, c.Request.Method)
		ext.HTTPUrl.Set(sp, c.Request.URL.String())
		ext.Component.Set(sp, "net/http")

		c.Request = c.Request.WithContext(opentracing2.ContextWithSpan(c.Request.Context(), sp))
		c.Next()

		ext.HTTPStatusCode.Set(sp, uint16(c.Writer.Status()))
		if c.Writer.Status() >= http.StatusInternalServerError {
			ext.Error.Set(sp, true)
		}
		sp.Finish()
	}
}

// GinTracerId If not found opentracing's tracer_id then generate a new tracer_id.
// Recommend to the end of the gin middleware.
func GinTracerID() gin.HandlerFunc {
	tracerID := tracerid.TracerID()
	return func(c *gin.Context) {
		sp := opentracing2.SpanFromContext(c.Request.Context())
		if sp == nil {
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(),
				tracerid.ActiveEsimKey, tracerID()))
		}

		c.Next()
	}
}

func GinMetaDataToCtx() gin.HandlerFunc {
	return func(c *gin.Context) {
		var meta = new(rctx.CommonParams)
		reqBuf, err := c.GetRawData()
		if err != nil {
			c.AbortWithStatus(http.StatusNotExtended)
			return
		}

		err = json.Unmarshal(reqBuf, meta)
		if err != nil {
			c.AbortWithStatus(http.StatusNotExtended)
			return
		}

		md := rctx.MD{
			rctx.LabelProdCd:    meta.ParseProdCd(),
			rctx.LabelAppID:     meta.AppID,
			rctx.LabelMerID:     meta.ParseMerID(),
			rctx.LabelRequestNo: meta.RequestNo,
			rctx.LabelTranCd:    meta.ParseTranCd(),
			rctx.LabelMethod:    c.Request.Method,
			rctx.LabelProtocol:  rctx.HTTPProtocol,
		}
		rctx.NewContext(c.Request.Context(), md)

		// request body put back to gin context body
		// MUST
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBuf))

		c.Next()
	}
}
