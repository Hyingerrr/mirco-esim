package redis

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/Hyingerrr/mirco-esim/config"
	"github.com/Hyingerrr/mirco-esim/container"
	logx "github.com/Hyingerrr/mirco-esim/log"
	"github.com/gomodule/redigo/redis"
)

var (
	poolOnce   sync.Once
	onceClient *Client
)

type Client struct {
	client *redis.Pool

	proxyNum int

	proxyInses []interface{}

	stateTicker time.Duration

	closeChan chan bool

	redisMaxActive int

	redisMaxIdle int

	redisIdleTimeout int

	redisHost string

	redisPort string

	redisPassword string

	redisReadTimeOut int64

	redisWriteTimeOut int64

	redisConnTimeOut int64

	dbIndex int // 默认0

	isTracer bool

	isMetric bool
}

type Option func(c *Client)

func NewClient(options ...Option) *Client {
	poolOnce.Do(func() {
		onceClient = &Client{
			stateTicker: 10 * time.Second,
			closeChan:   make(chan bool, 1),
		}

		for _, option := range options {
			option(onceClient)
		}

		onceClient.redisMaxActive = config.GetInt("redis_max_active")
		if onceClient.redisMaxActive == 0 {
			onceClient.redisMaxActive = 500
		}

		onceClient.redisMaxIdle = config.GetInt("redis_max_idle")
		if onceClient.redisMaxIdle == 0 {
			onceClient.redisMaxIdle = 100
		}

		onceClient.redisIdleTimeout = config.GetInt("redis_idle_time_out")
		if onceClient.redisIdleTimeout == 0 {
			onceClient.redisIdleTimeout = 600
		}

		onceClient.redisHost = config.GetString("redis_host")
		if onceClient.redisHost == "" {
			onceClient.redisHost = "0.0.0.0"
		}

		onceClient.redisPort = config.GetString("redis_port")
		if onceClient.redisPort == "" {
			onceClient.redisPort = "6379"
		}

		onceClient.redisPassword = config.GetString("redis_password")
		onceClient.dbIndex = config.GetInt("redis_db_index")

		onceClient.redisReadTimeOut = config.GetInt64("redis_read_time_out")
		if onceClient.redisReadTimeOut == 0 {
			onceClient.redisReadTimeOut = 300
		}

		onceClient.redisWriteTimeOut = config.GetInt64("redis_write_time_out")
		if onceClient.redisWriteTimeOut == 0 {
			onceClient.redisWriteTimeOut = 300
		}

		onceClient.redisConnTimeOut = config.GetInt64("redis_conn_time_out")
		if onceClient.redisConnTimeOut == 0 {
			onceClient.redisConnTimeOut = 300
		}

		onceClient.isTracer = config.GetBool("redis_tracer")
		onceClient.isMetric = config.GetBool("redis_metrics")

		onceClient.initPool()

		if config.GetString("runmode") == "pro" {
			// conn success ？
			rc := onceClient.client.Get()
			if rc.Err() != nil {
				logx.Panicf(rc.Err().Error())
			}
			rc.Close()
		}

		go onceClient.Stats()

		logx.Infof("[redis] init success %s : %s",
			onceClient.redisHost, onceClient.redisPort)
	})

	return onceClient
}

func WithStateTicker(stateTicker time.Duration) Option {
	return func(r *Client) {
		r.stateTicker = stateTicker
	}
}

// initClient Initialize the pool of connections.
func (c *Client) initPool() {
	c.client = &redis.Pool{
		MaxIdle:     c.redisMaxIdle,
		MaxActive:   c.redisMaxActive,
		IdleTimeout: time.Duration(c.redisIdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", c.redisHost+":"+c.redisPort,
				redis.DialReadTimeout(time.Duration(c.redisReadTimeOut)*time.Millisecond),
				redis.DialWriteTimeout(time.Duration(c.redisWriteTimeOut)*time.Millisecond),
				redis.DialConnectTimeout(time.Duration(c.redisConnTimeOut)*time.Millisecond))

			if err != nil {
				logx.Panicf("redis.Dial err: %s", err.Error())
				return nil, err
			}

			if c.redisPassword != "" {
				if _, err = conn.Do("AUTH", c.redisPassword); err != nil {
					err = conn.Close()
					logx.Panicf("redis.AUTH err: %s", err)
					return nil, err
				}
			}

			// select db
			_, err = conn.Do("SELECT", c.dbIndex)
			if err != nil {
				logx.Panicf("Select err: %s", err.Error())
				return nil, err
			}

			if config.GetBool("debug") {
				conn = redis.NewLoggingConn(
					conn, log.New(os.Stdout, "",
						log.Ldate|log.Ltime|log.Lshortfile), "")
			}
			return conn, nil
		},
	}
}

func (c *Client) GetRedisConn() redis.Conn {
	return c.client.Get()
}

func (c *Client) Close() error {
	err := c.client.Close()
	c.closeChan <- true

	return err
}

func (c *Client) Ping() error {
	conn := c.client.Get()

	return conn.Err()
}

func (c *Client) Stats() {
	ticker := time.NewTicker(c.stateTicker)
	var stats redis.PoolStats

	for {
		select {
		case <-ticker.C:
			stats = c.client.Stats()
			redisStats.Set(float64(stats.ActiveCount), []string{container.AppName(), "active_count"}...)
			redisStats.Set(float64(stats.IdleCount), []string{container.AppName(), "idle_count"}...)
		case <-c.closeChan:
			logx.Infof("stop stats")
			goto Stop
		}
	}
Stop:
	ticker.Stop()
}

func (c *Client) SubChannels(ctx context.Context,
	onStart func() error,
	onMessage func(channel string, data []byte) error,
	channels ...string) error {
	// A ping is set to the server with this period to test for the health of
	// the connection and server.
	const healthCheckPeriod = 20 * time.Second

	psc := redis.PubSubConn{Conn: c.GetRedisConn()}

	if err := psc.Subscribe(redis.Args{}.AddFlat(channels)...); err != nil {
		return err
	}

	done := make(chan error, 1)

	// Start a goroutine to receive notifications from the server.
	go func() {
		defer psc.Close()
		for {
			switch n := psc.Receive().(type) {
			case error:
				done <- n
				return
			case redis.Message:
				if err := onMessage(n.Channel, n.Data); err != nil {
					done <- err
					return
				}
			case redis.Subscription:
				switch n.Count {
				case len(channels):
					// Notify application when all channels are subscribed.
					if err := onStart(); err != nil {
						done <- err
						return
					}
				case 0:
					// Return from the goroutine when all channels are unsubscribed.
					done <- nil
					return
				}
			case redis.Pong:
				continue
			}

		}
	}()

	ticker := time.NewTicker(healthCheckPeriod)
	defer ticker.Stop()

	var err error
loop:
	for {
		select {
		case <-ticker.C:
			// Send ping to test health of connection and server. If
			// corresponding pong is not received, then receive on the
			// connection will timeout and the receive goroutine will exit.
			if err = psc.Ping(""); err != nil {
				break loop
			}
		case <-ctx.Done():
			break loop
		case err = <-done:
			// Return error from the receive goroutine.
			return err
		}
	}

	// Signal the receiving goroutine to exit by unsubscribing from all channels.
	psc.Unsubscribe()

	// Wait for goroutine to complete.
	return <-done
}
