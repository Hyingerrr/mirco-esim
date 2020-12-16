package mysql

import (
	"github.com/jukylin/esim/metrics"
)

var (
	mysqlTotal = metrics.CreateMetricCount("mysql_total", []string{"sql"}...)

	mysqlDuration = metrics.CreateMetricHistogram(
		"mysql_duration_seconds",
		[]float64{0.01, 0.05, 0.1, 0.5, 1},
		[]string{"sql"}...)

	mysqlStats = metrics.CreateMetricGauge("mysql_stats", []string{"db", "stats"}...)
)
