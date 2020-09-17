package metrics

import "github.com/prometheus/client_golang/prometheus"

// 汇总
type SummaryVec interface {
	Observe(v int64, labels ...string)
	close() bool
}

type summaryVec struct {
	summary *prometheus.SummaryVec
}

func NewSummaryVec(opts *SummaryOpts) *summaryVec {
	if opts == nil {
		return nil
	}

	vec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  opts.Namespace,
		Subsystem:  opts.Subsystem,
		Name:       opts.Name,
		Help:       opts.Help,
		Objectives: opts.Objectives,
	}, opts.Labels)
	prometheus.MustRegister(vec)
	return &summaryVec{summary: vec}
}

func (sv *summaryVec) Observe(v int64, labels ...string) {
	sv.summary.WithLabelValues(labels...).Observe(float64(v))
}

func (sv *summaryVec) close() bool {
	return prometheus.Unregister(sv.summary)
}
