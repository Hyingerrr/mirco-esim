package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGaugeSet(t *testing.T) {
	gaugeVec := NewGaugeVec(&GaugeVecOpts{
		Namespace: "test_prom",
		Subsystem: "request",
		Name:      "duration_ms",
		Help:      "http server requests duration(ms).",
		Labels:    []string{"tranCd"},
	})
	defer gaugeVec.close()
	gv, _ := gaugeVec.(*promGuageVec)

	//gv.Set(123, "M1020")

	gv.Add(11, "M1020")
	gv.Add(14, "M1020")
	r := testutil.ToFloat64(gv.gauge)
	//assert.Equal(t, float64(123), r)
	assert.Equal(t, float64(25), r)
}
