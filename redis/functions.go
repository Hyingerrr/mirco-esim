package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

// ------------------------------------------------------------ //
// --------------------------- STRING ------------------------- //
// ----------------------------------------------------------- //
// GET
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return redis.String(c.Do(ctx, "GET", key))
}

func (c *Client) GetBytes(ctx context.Context, key string) ([]byte, error) {
	return redis.Bytes(c.Do(ctx, "GET", key))
}

// SET, expiration 单位s, 0永久
func (c *Client) Set(ctx context.Context, key string, val interface{}, expiration int64) error {
	if expiration < 0 {
		_, err := c.Do(ctx, "SET", key, val)
		return err
	}

	_, err := c.Do(ctx, "SET", key, val, "EX", expiration)
	return err
}

// INCRBY
func (c *Client) IncrBy(ctx context.Context, key string, step int) (int, error) {
	return redis.Int(c.Do(ctx, "INCR", key, step))
}

// ------------------------------------------------------------ //
// --------------------------- HASH ------------------------- //
// ----------------------------------------------------------- //
func (c *Client) HSet(ctx context.Context, key string, field string, val interface{}) error {
	value, err := c.encode(val)
	if err != nil {
		return err
	}
	_, err = c.Do(ctx, "HSET", key, field, value)

	return err
}

func (c *Client) HMSet(ctx context.Context, key string, val interface{}) error {
	_, err := c.Do(ctx, "HMSET", redis.Args{}.Add(key).AddFlat(val)...)
	return err
}

func (c *Client) HGet(ctx context.Context, key, field string) ([]byte, error) {
	return redis.Bytes(c.Do(ctx, "HGET", key, field))
}

func (c *Client) HMGet(ctx context.Context, key string, field ...interface{}) ([]string, error) {
	return redis.Strings(c.Do(ctx, "HMGET", redis.Args{}.Add(key).AddFlat(field)...))
}

func (c *Client) HGetAll(ctx context.Context, key string, val interface{}) error {
	v, err := redis.Values(c.Do(ctx, "HGETALL", key))
	if err != nil {
		return err
	}

	return redis.ScanStruct(v, val)
}

// ------------------------------------------------------------ //
// --------------------------- LIST ------------------------- //
// ----------------------------------------------------------- //
// timeout 单位s， 0表示无限期阻塞
func (c *Client) BLPop(ctx context.Context, key string, timeout int) (interface{}, error) {
	values, err := redis.Values(c.Do(ctx, "BLPOP", key, timeout))
	if err != nil {
		return nil, err
	}

	if len(values) != 2 {
		return nil, fmt.Errorf("redisgo: unexpected number of values, got %d", len(values))
	}

	return values[1], err
}

func (c *Client) BRPop(ctx context.Context, key string, timeout int) (interface{}, error) {
	values, err := redis.Values(c.Do(ctx, "BRPOP", key, timeout))
	if err != nil {
		return nil, err
	}

	if len(values) != 2 {
		return nil, fmt.Errorf("redisgo: unexpected number of values, got %d", len(values))
	}

	return values[1], err
}

func (c *Client) LPop(ctx context.Context, key string) (interface{}, error) {
	return c.Do(ctx, "LPOP", key)
}

func (c *Client) RPop(ctx context.Context, key string) (interface{}, error) {
	return c.Do(ctx, "RPOP", key)
}

// LPush 将一个值插入到列表头部
func (c *Client) LPush(ctx context.Context, key string, val interface{}) error {
	value, err := c.encode(val)
	if err != nil {
		return err
	}

	_, err = c.Do(ctx, "LPUSH", key, value)

	return err
}

// RPush 将一个值插入到列表尾部
func (c *Client) RPush(ctx context.Context, key string, val interface{}) error {
	value, err := c.encode(val)
	if err != nil {
		return err
	}

	_, err = c.Do(ctx, "RPUSH", key, value)
	return err
}

// 区间以偏移量 start 和 end
// 0 表示列表的第一个元素， 1 表示列表的第二个元素，以此类推。
// -1 表示列表的最后一个元素， -2 表示列表的倒数第二个元素，以此类推。
// end (闭区间)
func (c *Client) LRange(ctx context.Context, key string, start, end int) (interface{}, error) {
	return c.Do(ctx, "LRANGE", key, start, end)
}

// ------------------------------------------------------------ //
// --------------------------- SET --------------------------- //
// ----------------------------------------------------------- //
func (c *Client) SAdd(ctx context.Context, key string, fields []interface{}) (bool, error) {
	return redis.Bool(c.Do(ctx, "SADD", redis.Args{}.Add(key).AddFlat(fields)...))
}

func (c *Client) SRem(ctx context.Context, key string, members []interface{}) (interface{}, error) {
	return c.Do(ctx, "SRem", redis.Args{}.Add(key).AddFlat(members)...)
}

func (c *Client) SMembers(ctx context.Context, key string) (interface{}, error) {
	return c.Do(ctx, "SMembers", key)
}

func (c *Client) SisMember(ctx context.Context, key string, member interface{}) (interface{}, error) {
	return c.Do(ctx, "SISMember", key, member)
}

// ------------------------------------------------------------ //
// ------------------------ SORTED SET  ----------------------- //
// ----------------------------------------------------------- //
func (c *Client) ZAdd(ctx context.Context, key string, score int64, member string) (interface{}, error) {
	return c.Do(ctx, "ZADD", key, score, member)
}

func (c *Client) ZRem(ctx context.Context, key string, member string) (interface{}, error) {
	return c.Do(ctx, "ZREM", key, member)
}

//  返回有序集成员member的score值。如果member元素不是有序集key的成员，或key不存在，返回nil 。
func (c *Client) ZScore(ctx context.Context, key string, member string) (int64, error) {
	return redis.Int64(c.Do(ctx, "ZSCORE", key, member))
}

// 返回有序集中指定成员的排名。其中有序集成员按分数值递增(从小到大)顺序排列。score 值最小的成员排名为 0
func (c *Client) ZRank(ctx context.Context, key, member string) (int64, error) {
	return redis.Int64(c.Do(ctx, "ZRANK", key, member))
}

// 返回有序集中成员的排名。其中有序集成员按分数值递减(从大到小)排序。分数值最大的成员排名为 0 。
func (c *Client) ZRevrank(ctx context.Context, key, member string) (int64, error) {
	return redis.Int64(c.Do(ctx, "ZREVRANK", key, member))
}

// ZRange 返回有序集中，指定区间内的成员。其中成员的位置按分数值递增(从小到大)来排序。具有相同分数值的成员按字典序来排列。
// 以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，以此类推。或 以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。
func (c *Client) ZRange(ctx context.Context, key string, from, to int64) (map[string]int64, error) {
	return redis.Int64Map(c.Do(ctx, "ZRANGE", key, from, to, "WITHSCORES"))
}

// ------------------------------------------------------------ //
// --------------------------- KEYS ------------------------- //
// ----------------------------------------------------------- //

// key 设置过期时间  单位s
func (c *Client) Expire(ctx context.Context, key string, expiration int64) error {
	_, err := redis.Bool(c.Do(ctx, "EXPIRE", key, expiration))

	return err
}

func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	return redis.Bool(c.Do(ctx, "EXISTS", key))
}

func (c *Client) DeleteKey(ctx context.Context, key string) (bool, error) {
	return redis.Bool(c.Do(ctx, "DEL", key))
}

func marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// encode 序列化要保存的值
func (c *Client) encode(val interface{}) (interface{}, error) {
	var value interface{}
	switch v := val.(type) {
	case string, int, uint, int8, int16, int32, int64, float32, float64, bool:
		value = v
	default:
		b, err := marshal(v)
		if err != nil {
			return nil, err
		}
		value = string(b)
	}
	return value, nil
}

func (c *Client) SelectDB(ctx context.Context, db int) error {
	_, err := c.Do(ctx, "SELECT", db)
	return err
}
