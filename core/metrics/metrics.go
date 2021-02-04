package metrics

import (
	"fmt"

	"github.com/pkg/errors"
)

const (
	_esimMetricNameSpace    = "esim"
	_esimSubsystemCount     = "count"
	_esimSubsystemGauge     = "gauge"
	_esimSubsystemHistogram = "histogram"
	_esimSubsystemSummary   = "summary"
)

var (
	_defaultBuckets = []float64{5, 10, 25, 50, 100, 250, 500}
)

func CreateMetricCount(name string, labels ...string) CounterVec {
	if name == "" || len(labels) == 0 {
		panic(errors.New("metric count name or labels should not be empty"))
	}

	return NewCounterVec(&CounterVecOpts{
		Namespace: _esimMetricNameSpace,
		Subsystem: _esimSubsystemCount,
		Name:      name,
		Labels:    labels,
		Help:      fmt.Sprintf("esim metric count %s", name),
	})
}

func CreateMetricGauge(name string, labels ...string) GaugeVec {
	if name == "" || len(labels) == 0 {
		panic(errors.New("metric gauge name or labels should not be empty"))
	}

	return NewGaugeVec(&GaugeVecOpts{
		Namespace: _esimMetricNameSpace,
		Subsystem: _esimSubsystemGauge,
		Name:      name,
		Labels:    labels,
		Help:      fmt.Sprintf("esim metrics gauge %s", name),
	})
}

func CreateMetricHistogram(name string, buckets []float64, labels ...string) HistogramVec {
	if name == "" || len(labels) == 0 {
		panic(errors.New("metric histogram name or labels should not be empty"))
	}

	if len(buckets) == 0 {
		buckets = _defaultBuckets
	}

	return NewHistogramVec(&HistogramVecOpts{
		Namespace: _esimMetricNameSpace,
		Subsystem: _esimSubsystemHistogram,
		Name:      name,
		Labels:    labels,
		Buckets:   buckets,
		Help:      fmt.Sprintf("esim metrics histogram %s", name),
	})
}

func CreateMetricSummary(name string, objectives map[float64]float64, labels ...string) SummaryVec {
	if name == "" || len(labels) == 0 {
		panic(errors.New("metric summary name or labels should not be empty"))
	}

	return NewSummaryVec(&SummaryOpts{
		Namespace:  _esimMetricNameSpace,
		Subsystem:  _esimSubsystemSummary,
		Name:       name,
		Help:       fmt.Sprintf("esim metrics histogram %s", name),
		Labels:     labels,
		Objectives: objectives,
	})
}
