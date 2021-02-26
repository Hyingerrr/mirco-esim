package mysql

import (
	"github.com/jukylin/esim/core/meta"
	"github.com/jukylin/esim/core/metrics"
)

var (
	mysqlDBCount = metrics.CreateMetricCount("mysql_total", []string{meta.ServiceName, "table"}...)

	mysqlDBDuration = metrics.CreateMetricHistogram(
		"mysql_duration_seconds",
		[]float64{0.01, 0.05, 0.1, 0.5, 1},
		[]string{meta.ServiceName, "table"}...)

	mysqlDBStats = metrics.CreateMetricGauge("mysql_stats", []string{meta.ServiceName, "db", "stats"}...)

	mysqlDBMiss = metrics.CreateMetricCount("mysql_miss", []string{meta.ServiceName, "schema", "table"}...)

	mysqlDBError = metrics.CreateMetricCount("mysql_error", []string{meta.ServiceName, "schema", "table"}...)
)
