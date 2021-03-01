package redis

import (
	"context"
	"fmt"
	"strconv"

	logx "github.com/jukylin/esim/log"

	"github.com/jukylin/esim/core/meta"

	"github.com/gomodule/redigo/redis"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func (c *Client) DoWithMetric(redisConn redis.Conn, cmd string, args ...interface{}) (interface{}, error) {
	redisCount.Inc(meta.ServiceName, cmd)
	reply, err := redisConn.Do(cmd, args...)
	if err != nil {
		redisErrCount.Inc(meta.ServiceName, cmd, args[0].(string))
	}
	return reply, err
}

func (c *Client) Do(ctx context.Context, command string, args ...interface{}) (reply interface{}, err error) {
	redisConn := c.GetRedisConn()
	defer redisConn.Close()

	if !c.isTracer {
		if c.isMetric {
			reply, err = c.DoWithMetric(redisConn, command, args...)
		} else {
			reply, err = redisConn.Do(command, args...)
		}
		if err != nil {
			logx.Errorc(ctx, "redis error:%v, key[%v]", err, args[0])
		}

		return
	}

	tc := c.withTrace(ctx)
	statement := getStatement(command, args...)
	span := tc.tracer.StartSpan(fmt.Sprintf("redis_%s", command), opentracing.ChildOf(tc.spanCtx))
	defer span.Finish()

	ext.DBType.Set(span, "redis")
	ext.DBInstance.Set(span, strconv.Itoa(tc.dbIndex))
	ext.PeerService.Set(span, "redis")
	ext.PeerHostname.Set(span, tc.redisHost)
	ext.SpanKindRPCClient.Set(span)

	if c.isMetric {
		reply, err = c.DoWithMetric(redisConn, command, args...)
	} else {
		reply, err = redisConn.Do(command, args...)
	}
	if err != nil {
		logx.Errorc(ctx, "redis error:%v, key[%v]", err, args[0])
		ext.DBStatement.Set(span, statement)
		if err != redis.ErrNil {
			ext.Error.Set(span, true)
			ext.MessageBusDestination.Set(span, err.Error())
			span.LogKV("event", "error", "message", err.Error())
		}
	} else {
		ext.DBStatement.Set(span, statement)
	}

	return reply, err
}
