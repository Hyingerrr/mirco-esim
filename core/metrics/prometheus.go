package metrics

import (
	"net/http"
	"strings"

	"github.com/Hyingerrr/mirco-esim/config"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const defaultPrometheusHTTPAddr = "9002"

type Prometheus struct{}

type VectorOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
}

type HistogramVecOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
	Buckets   []float64
}

type SummaryOpts struct {
	Namespace  string
	Subsystem  string
	Name       string
	Help       string
	Labels     []string
	Objectives map[float64]float64
}

func NewPrometheus() *Prometheus {
	addr := config.GetString("prometheus_http_addr")
	if addr == "" {
		addr = defaultPrometheusHTTPAddr
	}

	if in := strings.Index(addr, ":"); in < 0 {
		addr = ":" + addr
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			panic(err)
		}
	}()

	return &Prometheus{}
}
