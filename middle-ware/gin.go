package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	config2 "github.com/jukylin/esim/core/config"

	"github.com/jukylin/esim/pkg/common/meta"

	"github.com/gin-gonic/gin"
	tracerid "github.com/jukylin/esim/pkg/tracer-id"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func GinMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			start     = time.Now()
			ctx       = c.Request.Context()
			labels    []string
			respLabel []string
		)

		// request
		labels = append(labels,
			meta.String(ctx, meta.ServiceName),
			c.Request.URL.Path,
			meta.String(ctx, meta.TranCd),
			meta.String(ctx, meta.AppID))

		// response
		respLabel = append(respLabel,
			meta.String(ctx, meta.ServiceName),
			c.Request.URL.Path,
			meta.String(ctx, meta.TranCd),
			fmt.Sprintf("%d", c.Writer.Status()))

		c.Next()

		serverReqQPS.Inc(labels...)
		serverReqDuration.Observe(time.Since(start).Seconds(), labels...)
		responseStatus.Inc(respLabel...)
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

func GinMetaDataToCtx(conf config2.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var meta = new(meta.CommonParams)
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

		md := meta.MD{
			meta.LabelProdCd:      meta.ParseProdCd(),
			meta.LabelAppID:       meta.AppID,
			meta.LabelMerID:       meta.ParseMerID(),
			meta.LabelRequestNo:   meta.RequestNo,
			meta.LabelTranCd:      meta.ParseTranCd(),
			meta.LabelMethod:      c.Request.Method,
			meta.LabelProtocol:    meta.HTTPProtocol,
			meta.LabelUri:         c.Request.URL.Path,
			meta.LabelServiceName: conf.GetString("appname"),
		}
		rCtx := meta.NewContext(c.Request.Context(), md)
		c.Request = c.Request.WithContext(rCtx)

		// MUST: request body put back to gin context body
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBuf))

		c.Next()
	}
}
