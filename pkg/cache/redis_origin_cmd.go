package cache

import (
	"context"
	"time"

	"github.com/ishaqcherry9/depend/pkg/encoding"
	"github.com/redis/go-redis/v9"
)

type RedisOriginCmds interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	Decr(ctx context.Context, key string) *redis.IntCmd
	HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd
	HMGet(ctx context.Context, key string, field ...string) *redis.SliceCmd
	HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	HGet(ctx context.Context, key, field string) *redis.StringCmd
	HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd
	LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd
	LPop(ctx context.Context, key string) *redis.StringCmd
	RPop(ctx context.Context, key string) *redis.StringCmd
	SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
	ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd
	ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd
	ZIncrBy(ctx context.Context, key string, increment float64, member string) *redis.FloatCmd
	ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd
	Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	SetArgs(ctx context.Context, key string, value interface{}, a redis.SetArgs) *redis.StatusCmd
	Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error)
	Pipeline() redis.Pipeliner
	TxPipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error)
	TxPipeline() redis.Pipeliner
	Process(ctx context.Context, cmd redis.Cmder) error
	Options() *redis.Options
	PoolStats() *redis.PoolStats
	PSubscribe(ctx context.Context, channels ...string) *redis.PubSub
	SSubscribe(ctx context.Context, channels ...string) *redis.PubSub
	NewSearchBuilder(ctx context.Context, index, query string) *redis.SearchBuilder
	NewAggregateBuilder(ctx context.Context, index, query string) *redis.AggregateBuilder
	NewCreateIndexBuilder(ctx context.Context, index string) *redis.CreateIndexBuilder
	NewDropIndexBuilder(ctx context.Context, index string) *redis.DropIndexBuilder
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
}

type RedisOriginCmdCache struct {
	client            *redis.Client
	KeyPrefix         string
	encoding          encoding.Encoding
	DefaultExpireTime time.Duration
	newObject         func() interface{}
}

func NewRedisOriginCmdCache(client *redis.Client, keyPrefix string, encode encoding.Encoding, newObject func() interface{}) *RedisOriginCmdCache {
	return &RedisOriginCmdCache{
		client:    client,
		KeyPrefix: keyPrefix,
		encoding:  encode,
		newObject: newObject,
	}
}

func (c *RedisOriginCmdCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return c.client.Set(ctx, key, value, expiration)
}

func (c *RedisOriginCmdCache) Get(ctx context.Context, key string) *redis.StringCmd {
	return c.client.Get(ctx, key)
}

func (c *RedisOriginCmdCache) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	if len(keys) == 0 {
		return c.client.Del(ctx)
	}
	return c.client.Del(ctx, keys...)
}

func (c *RedisOriginCmdCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return c.client.SetNX(ctx, key, value, expiration)
}

func (c *RedisOriginCmdCache) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	if len(keys) == 0 {
		return c.client.Exists(ctx)
	}
	return c.client.Exists(ctx, keys...)
}

func (c *RedisOriginCmdCache) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return c.client.Expire(ctx, key, expiration)
}

func (c *RedisOriginCmdCache) Incr(ctx context.Context, key string) *redis.IntCmd {
	return c.client.Incr(ctx, key)
}

func (c *RedisOriginCmdCache) Decr(ctx context.Context, key string) *redis.IntCmd {
	return c.client.Decr(ctx, key)
}

func (c *RedisOriginCmdCache) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.client.HSet(ctx, key, values...)
}

func (c *RedisOriginCmdCache) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return c.client.HGet(ctx, key, field)
}

func (c *RedisOriginCmdCache) HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd {
	return c.client.HMSet(ctx, key, values...)
}

func (c *RedisOriginCmdCache) HMGet(ctx context.Context, key string, field ...string) *redis.SliceCmd {
	return c.client.HMGet(ctx, key, field...)
}

func (c *RedisOriginCmdCache) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	if len(fields) == 0 {
		return c.client.HDel(ctx, key)
	}
	return c.client.HDel(ctx, key, fields...)
}

func (c *RedisOriginCmdCache) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.client.LPush(ctx, key, values...)
}

func (c *RedisOriginCmdCache) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.client.RPush(ctx, key, values...)
}

func (c *RedisOriginCmdCache) LPop(ctx context.Context, key string) *redis.StringCmd {
	return c.client.LPop(ctx, key)
}

func (c *RedisOriginCmdCache) RPop(ctx context.Context, key string) *redis.StringCmd {
	return c.client.RPop(ctx, key)
}

func (c *RedisOriginCmdCache) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return c.client.SAdd(ctx, key, members...)
}

func (c *RedisOriginCmdCache) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	return c.client.SMembers(ctx, key)
}

func (c *RedisOriginCmdCache) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	return c.client.ZAdd(ctx, key, members...)
}

func (c *RedisOriginCmdCache) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return c.client.ZRange(ctx, key, start, stop)
}

func (c *RedisOriginCmdCache) ZRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	return c.client.ZRangeWithScores(ctx, key, start, stop)
}

func (c *RedisOriginCmdCache) ZIncrBy(ctx context.Context, key string, increment float64, member string) *redis.FloatCmd {
	return c.client.ZIncrBy(ctx, key, increment, member)
}

func (c *RedisOriginCmdCache) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return c.client.ZRem(ctx, key, members...)
}

func (c *RedisOriginCmdCache) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	return c.client.Publish(ctx, channel, message)
}

func (c *RedisOriginCmdCache) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.client.Subscribe(ctx)
	}
	return c.client.Subscribe(ctx, channels...)
}

func (c *RedisOriginCmdCache) SetArgs(ctx context.Context, key string, value interface{}, a redis.SetArgs) *redis.StatusCmd {
	return c.client.SetArgs(ctx, key, value, a)
}

func (c *RedisOriginCmdCache) Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return c.client.Pipelined(ctx, fn)
}

func (c *RedisOriginCmdCache) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

func (c *RedisOriginCmdCache) TxPipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return c.client.TxPipelined(ctx, fn)
}

func (c *RedisOriginCmdCache) TxPipeline() redis.Pipeliner {
	return c.client.TxPipeline()
}

func (c *RedisOriginCmdCache) Process(ctx context.Context, cmd redis.Cmder) error {
	return c.client.Process(ctx, cmd)
}

func (c *RedisOriginCmdCache) Options() *redis.Options {
	return c.client.Options()
}

func (c *RedisOriginCmdCache) PoolStats() *redis.PoolStats {
	return c.client.PoolStats()
}

func (c *RedisOriginCmdCache) PSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.client.PSubscribe(ctx)
	}
	return c.client.PSubscribe(ctx, channels...)
}

func (c *RedisOriginCmdCache) SSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.client.SSubscribe(ctx)
	}
	return c.client.SSubscribe(ctx, channels...)
}

func (c *RedisOriginCmdCache) NewSearchBuilder(ctx context.Context, index, query string) *redis.SearchBuilder {
	return c.client.NewSearchBuilder(ctx, index, query)
}

func (c *RedisOriginCmdCache) NewAggregateBuilder(ctx context.Context, index, query string) *redis.AggregateBuilder {
	return c.client.NewAggregateBuilder(ctx, index, query)
}

func (c *RedisOriginCmdCache) NewCreateIndexBuilder(ctx context.Context, index string) *redis.CreateIndexBuilder {
	return c.client.NewCreateIndexBuilder(ctx, index)
}

func (c *RedisOriginCmdCache) NewDropIndexBuilder(ctx context.Context, index string) *redis.DropIndexBuilder {
	return c.client.NewDropIndexBuilder(ctx, index)
}

// RedisClient 返回底层原始的Client，用于需要直接访问底层 API 的特殊场景
func (c *RedisOriginCmdCache) RedisClient() *redis.Client {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client
}

// Eval 允许执行 Lua 脚本（EVAL）——传入的 keys 会被包装器加上前缀。
func (c *RedisOriginCmdCache) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	if len(keys) == 0 {
		return c.client.Eval(ctx, script, nil, args...)
	}
	return c.client.Eval(ctx, script, keys, args...)
}
