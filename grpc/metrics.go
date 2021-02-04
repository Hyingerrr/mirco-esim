package grpc

import (
	"github.com/jukylin/esim/core/meta"
	"github.com/jukylin/esim/core/metrics"
)

var (
	_serverGRPCReqQPS = metrics.CreateMetricCount(
		"grpc_server_requests_QPS",
		[]string{meta.ServiceName, meta.Uri, meta.AppID}...,
	)

	_serverGRPCReqDuration = metrics.CreateMetricHistogram(
		"grpc_server_requests_duration_ms",
		[]float64{5, 10, 25, 50, 100, 250, 500, 1000, 2000},
		[]string{meta.ServiceName, meta.Uri, meta.AppID}...,
	)

	_clientGRPCReqQPS = metrics.CreateMetricCount(
		"grpc_client_requests_QPS",
		[]string{meta.ServiceName, meta.Uri, meta.AppID, meta.StatusCode}...,
	)

	_clientGRPCReqDuration = metrics.CreateMetricHistogram(
		"grpc_client_requests_duration_ms",
		[]float64{5, 10, 25, 50, 100, 250, 500, 1000, 2000},
		[]string{meta.ServiceName, meta.Uri, meta.AppID}...,
	)
)
