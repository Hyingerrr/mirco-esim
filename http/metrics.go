package http

import (
	"github.com/jukylin/esim/core/meta"
	"github.com/jukylin/esim/core/metrics"
)

var (
	httpCallReqError = metrics.CreateMetricCount(
		"http_call_resp_error",
		[]string{meta.ServiceName, meta.Uri}...,
	)

	httpCallRespCount = metrics.CreateMetricCount(
		"http_call_resp",
		[]string{meta.ServiceName, meta.Uri, meta.StatusCode}...,
	)

	httpCallReqDuration = metrics.CreateMetricHistogram(
		"http_call_dms",
		[]float64{0.02, 0.08, 0.15, 0.5, 1, 3},
		[]string{meta.ServiceName, meta.Uri}...,
	)
)
