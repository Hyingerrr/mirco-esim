package metrics

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestHistogramObserve(t *testing.T) {
	histogramVec := NewHistogramVec(&HistogramVecOpts{
		Name:    "counts",
		Help:    "test server requests duration(ms).",
		Buckets: []float64{1, 2, 3},
		Labels:  []string{"tranCd"},
	})
	defer histogramVec.close()
	hv, _ := histogramVec.(*promHistogramVec)
	hv.Observe(2, "M1010")

	metadata := `
		# HELP counts test server requests duration(ms).
        # TYPE counts histogram
`
	val := `
		counts_bucket{tranCd="M1010",le="1"} 0
		counts_bucket{tranCd="M1010",le="2"} 1
		counts_bucket{tranCd="M1010",le="3"} 1
		counts_bucket{tranCd="M1010",le="+Inf"} 1
		counts_sum{tranCd="M1010"} 2
        counts_count{tranCd="M1010"} 1
`

	err := testutil.CollectAndCompare(hv.histogram, strings.NewReader(metadata+val))
	assert.Nil(t, err)
}
