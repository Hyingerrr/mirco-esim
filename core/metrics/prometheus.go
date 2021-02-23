package metrics

import (
	"net/http"
	"strings"

	"github.com/jukylin/esim/config"
	logx "github.com/jukylin/esim/log"
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
		logx.Panicf(http.ListenAndServe(addr, nil).Error())
	}()

	logx.Info("Prometheus Server Init Success")

	return &Prometheus{}
}
