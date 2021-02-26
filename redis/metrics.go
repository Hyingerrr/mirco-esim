package redis

import (
	"github.com/jukylin/esim/core/meta"
	"github.com/jukylin/esim/core/metrics"
)

var (
	redisErrCount = metrics.CreateMetricCount("redis_error", []string{meta.ServiceName, "cmd", "key"}...)
	redisCount    = metrics.CreateMetricCount("redis_count", []string{meta.ServiceName, "cmd"}...)
	redisStats    = metrics.CreateMetricGauge("redis_stats", []string{meta.ServiceName, "stats"}...)
)
