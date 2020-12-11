package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// 直方图
type HistogramVec interface {
	Observe(v float64, labels ...string)
	close() bool
}

type promHistogramVec struct {
	histogram *prometheus.HistogramVec
}

func NewHistogramVec(opts *HistogramVecOpts) HistogramVec {
	if opts == nil {
		return nil
	}

	vec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
		Buckets:   opts.Buckets,
	}, opts.Labels)
	prometheus.MustRegister(vec)
	return &promHistogramVec{histogram: vec}
}

func (hv *promHistogramVec) Observe(v float64, labels ...string) {
	hv.histogram.WithLabelValues(labels...).Observe(v)
}

func (hv *promHistogramVec) close() bool {
	return prometheus.Unregister(hv.histogram)
}
