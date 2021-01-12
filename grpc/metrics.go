package grpc

import (
	"github.com/jukylin/esim/core/meta"
	"github.com/jukylin/esim/metrics"
)

var (
	ServerGRPCReqQPS = metrics.CreateMetricCount(
		"grpc_server_requests_QPS",
		[]string{meta.ServiceName, meta.Uri, meta.AppID}...,
	)

	ServerGRPCReqDuration = metrics.CreateMetricHistogram(
		"grpc_server_requests_duration_ms",
		[]float64{5, 10, 25, 50, 100, 250, 500, 1000, 2000},
		[]string{meta.ServiceName, meta.Uri, meta.AppID}...,
	)
)
