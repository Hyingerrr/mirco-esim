package middleware

import (
	"github.com/jukylin/esim/metrics"
	"github.com/jukylin/esim/pkg/common/meta"
)

// request_total.
var serverReqQPS = metrics.CreateMetricCount(
	"requests_QPS",
	[]string{meta.ServiceName, meta.Uri, meta.TranCd, meta.AppID}...)

// request_duration_seconds.
var serverReqDuration = metrics.CreateMetricHistogram(
	"requests_duration_seconds",
	[]float64{0.1, 0.3, 0.5, 0.7, 0.9, 1, 3, 5, 10, 30, 100},
	[]string{meta.ServiceName, meta.Uri, meta.TranCd, meta.AppID}...)

// response_status_stats
var responseStatus = metrics.CreateMetricCount(
	"response_status",
	[]string{meta.ServiceName, meta.Uri, meta.TranCd, "status"}...)
