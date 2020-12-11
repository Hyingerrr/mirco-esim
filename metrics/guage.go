package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type GaugeVecOpts VectorOpts

// 测量仪
type GaugeVec interface {
	Set(v float64, labels ...string)
	Inc(labels ...string)
	Add(v float64, labels ...string)
	close() bool
}

type promGuageVec struct {
	gauge *prometheus.GaugeVec
}

func NewGaugeVec(opts *GaugeVecOpts) GaugeVec {
	if opts == nil {
		return nil
	}

	vec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: opts.Namespace,
			Subsystem: opts.Subsystem,
			Name:      opts.Name,
			Help:      opts.Help,
		}, opts.Labels)
	prometheus.MustRegister(vec)

	return &promGuageVec{gauge: vec}
}

func (gv *promGuageVec) Inc(labels ...string) {
	gv.gauge.WithLabelValues(labels...).Inc()
}

func (gv *promGuageVec) Add(v float64, labels ...string) {
	gv.gauge.WithLabelValues(labels...).Add(v)
}

func (gv *promGuageVec) Set(v float64, labels ...string) {
	gv.gauge.WithLabelValues(labels...).Set(v)
}

func (gv *promGuageVec) close() bool {
	return prometheus.Unregister(gv.gauge)
}
