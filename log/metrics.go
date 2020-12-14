package log

import (
	"runtime"
	"strconv"

	"github.com/jukylin/esim/metrics"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logMetricErrorCounter = metrics.CreateMetricCount("log_error_stats", "caller")

func addErrMetric(entry zapcore.Entry) error {
	if entry.Level == zap.ErrorLevel {
		logMetricErrorCounter.Inc(funcName(3))
	}

	return nil
}

func funcName(skip int) (name string) {
	if _, file, lineNo, ok := runtime.Caller(skip); ok {
		return file + ":" + strconv.Itoa(lineNo)
	}
	return "unknown:0"
}
