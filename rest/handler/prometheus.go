package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jukylin/esim/core/meta"
	"github.com/jukylin/esim/metrics"
)

// request_total.
var serverReqQPS = metrics.CreateMetricCount(
	"http_requests_QPS",
	[]string{meta.ServiceName, meta.Uri, meta.TranCd, meta.AppID}...)

// request_duration_seconds.
var serverReqDuration = metrics.CreateMetricHistogram(
	"http_requests_duration_seconds",
	[]float64{0.1, 0.3, 0.5, 0.7, 0.9, 1, 3, 5, 10, 30, 100},
	[]string{meta.ServiceName, meta.Uri, meta.TranCd, meta.AppID}...)

// response_status_stats
var responseStatus = metrics.CreateMetricCount(
	"http_response_status",
	[]string{meta.ServiceName, meta.Uri, meta.TranCd, "status"}...)

func HttpMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			start   = time.Now()
			labels  []string
			rplabel []string
		)

		var getCtx = func(label string) string {
			return meta.String(c.Request.Context(), label)
		}

		// request
		labels = append(labels, getCtx(meta.ServiceName), c.Request.URL.Path, getCtx(meta.TranCd), getCtx(meta.AppID))

		// response
		rplabel = append(rplabel, getCtx(meta.ServiceName), c.Request.URL.Path, getCtx(meta.TranCd), fmt.Sprintf("%d", c.Writer.Status()))

		serverReqQPS.Inc(labels...)
		serverReqDuration.Observe(time.Since(start).Seconds(), labels...)

		c.Next()

		responseStatus.Inc(rplabel...)
	}
}
