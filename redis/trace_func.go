package redis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gomodule/redigo/redis"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func (c *Client) Do(ctx context.Context, command string, args ...interface{}) (interface{}, error) {
	redisConn := c.GetCtxRedisConn()
	defer redisConn.Close()

	tc := c.withTrace(ctx)
	statement := getStatement(command, args...)
	span := tc.tracer.StartSpan(fmt.Sprintf("redis_%s", command), opentracing.ChildOf(tc.spanCtx))
	defer span.Finish()

	ext.DBType.Set(span, "redis")
	ext.DBInstance.Set(span, strconv.Itoa(tc.dbIndex))
	ext.PeerService.Set(span, "redis")
	ext.PeerHostname.Set(span, tc.redisHost)
	ext.SpanKindRPCClient.Set(span)

	reply, err := redisConn.Do(ctx, command, args...)
	if err != nil {
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
