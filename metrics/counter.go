package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type CounterVecOpts VectorOpts

// 计数器
type CounterVec interface {
	Inc(labels ...string)
	Add(v float64, labels ...string)
	GetMetric(labels ...string) (prometheus.Counter, error)
	close() bool
}

type promCounterVec struct {
	counter *prometheus.CounterVec
}

func NewCounterVec(opts *CounterVecOpts) CounterVec {
	if opts == nil {
		return nil
	}

	vec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
	}, opts.Labels)

	prometheus.MustRegister(vec)

	return &promCounterVec{counter: vec}
}

// +1
func (cv *promCounterVec) Inc(labels ...string) {
	cv.counter.WithLabelValues(labels...).Inc()
}

// +n
func (cv *promCounterVec) Add(n float64, labels ...string) {
	cv.counter.WithLabelValues(labels...).Add(n)
}

// close
func (cv *promCounterVec) close() bool {
	return prometheus.Unregister(cv.counter)
}

func (cv *promCounterVec) GetMetric(labels ...string) (prometheus.Counter, error) {
	return cv.counter.GetMetricWithLabelValues(labels...)
}
