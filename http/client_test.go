package http

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"

	"fmt"

	"github.com/Hyingerrr/mirco-esim/config"
	"github.com/Hyingerrr/mirco-esim/log"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

var logger log.Logger

const (
	host1 = "http://192.168.3.154:8081/ping"

	host2 = "127.0.0.2"
)

func TestMain(m *testing.M) {
	loggerOptions := log.LoggerOptions{}
	options := config.ViperConfOptions{}
	conf := config.NewViperConfig(options.WithConfigType("yaml"),
		options.WithConfFile([]string{"/Users/hy/develop/esim/config/a.yaml"}))
	logger = log.NewLogger(loggerOptions.WithDebug(true), loggerOptions.WithLoggerConf(conf))

	code := m.Run()

	os.Exit(code)
}

//nolint:dupl
func TestMulLevelRoundTrip(t *testing.T) {
	httpClient := NewClient()
	testCases := []struct {
		behavior string
		url      string
		result   int
	}{
		{fmt.Sprintf("%s:%d", host1, http.StatusOK), host1, http.StatusOK},
		{fmt.Sprintf("%s:%d", host2, http.StatusMultipleChoices), host2, http.StatusMultipleChoices},
	}

	ctx := context.Background()
	for _, test := range testCases {
		test := test
		t.Run(test.behavior, func(t *testing.T) {
			resp, err := httpClient.Get(ctx, test.url)
			resp.Body.Close()

			assert.Nil(t, err)
			assert.Equal(t, test.result, resp.StatusCode)
		})
	}
}

//nolint:dupl
func TestMonitorProxy(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("http_client_metrics", true)

	httpClient := NewClient()

	ctx := context.Background()
	resp, err := httpClient.Get(ctx, host1)
	assert.Nil(t, err)
	//resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	fmt.Println(string(buf))
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	rtyResp, err := httpClient.SendGet(ctx, host1)
	assert.Nil(t, err)
	fmt.Println(rtyResp.Body())
	assert.Equal(t, resp.StatusCode, http.StatusMultipleChoices)

	lab := prometheus.Labels{"url": host1, "method": http.MethodGet}
	c, _ := httpTotal.GetMetricWith(lab)
	metric := &io_prometheus_client.Metric{}
	err = c.Write(metric)
	assert.Nil(t, err)

	assert.Equal(t, float64(1), metric.Counter.GetValue())
}

//nolint:dupl
func TestTimeoutProxy(t *testing.T) {
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)
	memConfig.Set("http_client_check_slow", true)
	memConfig.Set("http_client_slow_time", 5)

	httpClient := NewClient()

	ctx := context.Background()
	resp, err := httpClient.Get(ctx, host1)
	//resp.Body.Close()

	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func TestClient_Post(t *testing.T) {
	var (
		it  = assert.New(t)
		url = "https://notify-test.eycard.cn:7443/WorthTech_Access_AppPaySystemV2/apppayacc"
		req = "channelid=D01X20200424011&merid=831290456990006&notifymobileno=18256083885&notifyusername=HY" +
			"&opt=zwrefund&oriwtorderid=11420200716190044117038&sign=A86CE990D5EA4A2EBBCA1E476C9F0&termid=" +
			"32765486&tradeamt=1&tradetrace=2020071728709821431677759"
		beg = time.Now()
	)

	fmt.Printf("start time: %v\n", beg.String())

	httpClient := NewClient()
	httpClient.RC().OnAfterResponse(func(client *resty.Client, response *resty.Response) error {
		fmt.Printf("end time: %v\n", time.Now().String())
		fmt.Printf("cost: %v\n", time.Since(beg).String())
		fmt.Println(response.Request.RawRequest.URL.Path)
		fmt.Println("get result: " + string(response.Body()))

		return nil
	})
	resp, err := httpClient.Post(context.Background(), url, "application/x-www-form-urlencoded;charset=UTF-8", strings.NewReader(req))
	it.Nil(err)
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	it.Nil(err)
	it.Equal(200, resp.StatusCode)
	fmt.Println(string(buf))
	fmt.Printf("resp: %v", resp)
}
