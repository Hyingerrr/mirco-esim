package mysql

import (
	"github.com/jukylin/esim/metrics"
)

var (
	mysqlDBTotal = metrics.CreateMetricCount("mysql_total", []string{"sql"}...)

	mysqlDBDuration = metrics.CreateMetricHistogram(
		"mysql_duration_seconds",
		[]float64{0.01, 0.05, 0.1, 0.5, 1},
		[]string{"sql"}...)

	mysqlDBStats = metrics.CreateMetricGauge("mysql_stats", []string{"db", "stats"}...)

	mysqlDBMiss = metrics.CreateMetricCount("mysql_miss", []string{"schema", "table"}...)

	mysqlDBError = metrics.CreateMetricCount("mysql_error", []string{"schema", "table"}...)
)
