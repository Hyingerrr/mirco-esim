package redis

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	//logger := log.NewLogger()

	//pool, err := dockertest.NewPool("")
	//if err != nil {
	//	logger.Fatalf("Could not connect to docker : %s", err)
	//}
	//opt := &dockertest.RunOptions{
	//	Repository: "redis",
	//	Tag:        "latest",
	//}
	//
	//resource, err := pool.RunWithOptions(opt, func(hostConfig *dc.HostConfig) {
	//	hostConfig.PortBindings = map[dc.Port][]dc.PortBinding{
	//		"6379/tcp": {{HostIP: "", HostPort: "6379"}},
	//	}
	//})
	//if err != nil {
	//	logger.Fatalf("Could not start resource: %s", err.Error())
	//}
	//
	//err = resource.Expire(60)
	//if err != nil {
	//	logger.Fatalf(err.Error())
	//}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	//if err := pool.Purge(resource); err != nil {
	//	logger.Fatalf("Could not purge resource: %s", err)
	//}

	os.Exit(code)
}

func TestGetProxyConn(t *testing.T) {
	poolOnce = sync.Once{}
	redisClientOptions := ClientOptions{}
	redisClent := NewClient(
		redisClientOptions.WithProxy(
			func() interface{} {
				monitorProxyOptions := MonitorProxyOptions{}
				return NewMonitorProxy(
					monitorProxyOptions.WithLogger(log.NewLogger()),
				)
			},
		),
	)

	conn := redisClent.GetCtxRedisConn()
	assert.IsTypef(t, NewMonitorProxy(), conn, "MonitorProxy type")
	assert.NotNil(t, conn)
	err := conn.Close()
	assert.Nil(t, err)

	err = redisClent.Close()
	assert.Nil(t, err)
}

func TestGetNotProxyConn(t *testing.T) {
	poolOnce = sync.Once{}
	redisClientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	redisClent := NewClient(
		redisClientOptions.WithConf(memConfig),
		redisClientOptions.WithProxy(
			func() interface{} {
				monitorProxyOptions := MonitorProxyOptions{}
				return NewMonitorProxy(
					monitorProxyOptions.WithLogger(log.NewLogger()),
				)
			},
		),
	)

	conn := redisClent.GetCtxRedisConn()

	assert.IsTypef(t, NewMonitorProxy(), conn, "MonitorProxy type")
	assert.NotNil(t, conn)
	err := conn.Close()
	assert.Nil(t, err)

	err = redisClent.Close()
	assert.Nil(t, err)
}

func TestMonitorProxy_Do(t *testing.T) {
	poolOnce = sync.Once{}
	redisClientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	redisClent := NewClient(
		redisClientOptions.WithConf(memConfig),
		redisClientOptions.WithProxy(
			func() interface{} {
				monitorProxyOptions := MonitorProxyOptions{}
				return NewMonitorProxy(
					monitorProxyOptions.WithLogger(log.NewLogger()),
				)
			},
		),
	)

	ctx := context.Background()

	conn := redisClent.GetCtxRedisConn()

	_, err := String(conn.Do(ctx, "get", "name"))
	assert.Nil(t, err)
	err = conn.Close()
	assert.Nil(t, err)

	err = redisClent.Close()
	assert.Nil(t, err)
}

func TestMulLevelProxy_Do(t *testing.T) {
	poolOnce = sync.Once{}
	redisClientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	spyProxy := newSpyProxy(log.NewLogger(), "spyProxy")

	redisClent := NewClient(
		redisClientOptions.WithConf(memConfig),
		redisClientOptions.WithProxy(
			func() interface{} {
				monitorProxyOptions := MonitorProxyOptions{}
				return NewMonitorProxy(
					monitorProxyOptions.WithLogger(log.NewLogger()),
				)
			},
			func() interface{} {
				return spyProxy
			},
		),
	)

	ctx := context.Background()

	conn := redisClent.GetCtxRedisConn()
	_, err := conn.Do(ctx, "get", "name")
	assert.Nil(t, err)

	assert.True(t, spyProxy.DoWasCalled)
	err = conn.Close()
	assert.Nil(t, err)

	err = redisClent.Close()
	assert.Nil(t, err)
}

func Benchmark_MulGo_Do(b *testing.B) {
	poolOnce = sync.Once{}
	redisClientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	redisClent := NewClient(
		redisClientOptions.WithConf(memConfig),
	)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wg := sync.WaitGroup{}
			ctx := context.Background()
			wg.Add(b.N * 2)

			testFunc := func(args, expected string) {
				conn := redisClent.GetCtxRedisConn()
				name, err := String(conn.Do(ctx, "get", args))
				assert.Nil(b, err)
				assert.Equal(b, expected, name)

				err = conn.Close()
				assert.Nil(b, err)
				wg.Done()
			}

			for j := 0; j < b.N; j++ {
				go func() {
					testFunc("name", "test")
				}()

				go func() {
					testFunc("version", "2.0")
				}()
			}
			wg.Wait()
		}
	})

	err := redisClent.Close()
	assert.Nil(b, err)
}

func TestRedisClient_Stats(t *testing.T) {
	poolOnce = sync.Once{}
	redisClientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	redisClent := NewClient(
		redisClientOptions.WithConf(memConfig),
		redisClientOptions.WithStateTicker(10*time.Microsecond),
	)

	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()
	conn := redisClent.GetCtxRedisConn()
	_, err := conn.Do(ctx, "get", "name")
	assert.Nil(t, err)

	_, err = conn.Do(ctx, "get", "name")
	assert.Nil(t, err)

	_, err = conn.Do(ctx, "get", "name")
	assert.Nil(t, err)
	err = conn.Close()
	assert.Nil(t, err)

	lab := prometheus.Labels{"stats": "active_count"}
	c, _ := redisStats.GetMetricWith(lab)
	metric := &io_prometheus_client.Metric{}
	err = c.Write(metric)
	assert.Nil(t, err)

	assert.True(t, metric.Gauge.GetValue() >= 0)

	err = redisClent.Close()
	assert.Nil(t, err)
}
