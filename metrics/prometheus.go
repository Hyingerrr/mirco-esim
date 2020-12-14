package metrics

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
