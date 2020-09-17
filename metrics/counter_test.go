package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNewCounterVec(t *testing.T) {
	var (
		it = assert.New(t)
	)

	c := NewCounterVec(&CounterVecOpts{
		Namespace: "test_server",
		Subsystem: "requests",
		Name:      "total",
		Help:      "test server total count",
	})
	defer c.close()
	it.NotNil(c)
}

func TestCounter(t *testing.T) {
	counterVec := NewCounterVec(&CounterVecOpts{
		Namespace: "test_server",
		Subsystem: "call",
		Name:      "code_total",
		Help:      "test server requests error count.",
		Labels:    []string{"tranCd", "code"},
	})
	defer counterVec.close()
	c, _ := counterVec.(*promCounterVec)
	//c.Inc("M1020", "500")
	//c.Inc("M1020", "500")

	c.Add(11, "M1021", "404")
	c.Add(12, "M1021", "404")
	r := testutil.ToFloat64(c.counter)
	//assert.Equal(t, float64(2), r)
	assert.Equal(t, float64(23), r)
}
