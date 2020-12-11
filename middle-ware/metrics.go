package middleware

import (
	"github.com/jukylin/esim/metrics"
	"github.com/jukylin/esim/pkg/common/rctx"
)

// request_total.
var serverReqQPS = metrics.CreateMetricCount(
	"requests_QPS",
	[]string{rctx.LabelEndpoint, rctx.LabelTranCd, rctx.LabelProdCd, rctx.LabelAppID}...)

// request_duration_seconds.
var serverReqDuration = metrics.CreateMetricHistogram(
	"requests_duration_seconds",
	[]float64{0.1, 0.3, 0.5, 0.7, 0.9, 1, 3, 5, 10, 30, 100},
	[]string{rctx.LabelEndpoint, rctx.LabelTranCd, rctx.LabelProdCd, rctx.LabelAppID}...)
