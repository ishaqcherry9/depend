package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type OriginCmds struct {
	Client *redis.Client
}

func (c *OriginCmds) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return c.Client.Set(ctx, key, value, expiration)
}

func (c *OriginCmds) Get(ctx context.Context, key string) *redis.StringCmd {
	return c.Client.Get(ctx, key)
}

func (c *OriginCmds) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	if len(keys) == 0 {
		return c.Client.Del(ctx)
	}
	return c.Client.Del(ctx, keys...)
}

func (c *OriginCmds) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return c.Client.SetNX(ctx, key, value, expiration)
}

func (c *OriginCmds) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	if len(keys) == 0 {
		return c.Client.Exists(ctx)
	}
	return c.Client.Exists(ctx, keys...)
}

func (c *OriginCmds) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return c.Client.Expire(ctx, key, expiration)
}

func (c *OriginCmds) Incr(ctx context.Context, key string) *redis.IntCmd {
	return c.Client.Incr(ctx, key)
}

func (c *OriginCmds) Decr(ctx context.Context, key string) *redis.IntCmd {
	return c.Client.Decr(ctx, key)
}

func (c *OriginCmds) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.Client.HSet(ctx, key, values...)
}

func (c *OriginCmds) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return c.Client.HGet(ctx, key, field)
}

func (c *OriginCmds) HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd {
	return c.Client.HMSet(ctx, key, values...)
}

func (c *OriginCmds) HMGet(ctx context.Context, key string, field ...string) *redis.SliceCmd {
	return c.Client.HMGet(ctx, key, field...)
}

func (c *OriginCmds) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	if len(fields) == 0 {
		return c.Client.HDel(ctx, key)
	}
	return c.Client.HDel(ctx, key, fields...)
}

func (c *OriginCmds) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.Client.LPush(ctx, key, values...)
}

func (c *OriginCmds) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.Client.RPush(ctx, key, values...)
}

func (c *OriginCmds) LPop(ctx context.Context, key string) *redis.StringCmd {
	return c.Client.LPop(ctx, key)
}

func (c *OriginCmds) RPop(ctx context.Context, key string) *redis.StringCmd {
	return c.Client.RPop(ctx, key)
}

func (c *OriginCmds) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return c.Client.SAdd(ctx, key, members...)
}

func (c *OriginCmds) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	return c.Client.SMembers(ctx, key)
}

func (c *OriginCmds) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	return c.Client.ZAdd(ctx, key, members...)
}

func (c *OriginCmds) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return c.Client.ZRange(ctx, key, start, stop)
}

func (c *OriginCmds) ZRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	return c.Client.ZRangeWithScores(ctx, key, start, stop)
}

func (c *OriginCmds) ZIncrBy(ctx context.Context, key string, increment float64, member string) *redis.FloatCmd {
	return c.Client.ZIncrBy(ctx, key, increment, member)
}

func (c *OriginCmds) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return c.Client.ZRem(ctx, key, members...)
}

func (c *OriginCmds) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	return c.Client.Publish(ctx, channel, message)
}

func (c *OriginCmds) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.Client.Subscribe(ctx)
	}
	return c.Client.Subscribe(ctx, channels...)
}

func (c *OriginCmds) SetArgs(ctx context.Context, key string, value interface{}, a redis.SetArgs) *redis.StatusCmd {
	return c.Client.SetArgs(ctx, key, value, a)
}

func (c *OriginCmds) Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return c.Client.Pipelined(ctx, fn)
}

func (c *OriginCmds) Pipeline() redis.Pipeliner {
	return c.Client.Pipeline()
}

func (c *OriginCmds) TxPipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return c.Client.TxPipelined(ctx, fn)
}

func (c *OriginCmds) TxPipeline() redis.Pipeliner {
	return c.Client.TxPipeline()
}

func (c *OriginCmds) Process(ctx context.Context, cmd redis.Cmder) error {
	return c.Client.Process(ctx, cmd)
}

func (c *OriginCmds) Options() *redis.Options {
	return c.Client.Options()
}

func (c *OriginCmds) PoolStats() *redis.PoolStats {
	return c.Client.PoolStats()
}

func (c *OriginCmds) PSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.Client.PSubscribe(ctx)
	}
	return c.Client.PSubscribe(ctx, channels...)
}

func (c *OriginCmds) SSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.Client.SSubscribe(ctx)
	}
	return c.Client.SSubscribe(ctx, channels...)
}

func (c *OriginCmds) NewSearchBuilder(ctx context.Context, index, query string) *redis.SearchBuilder {
	return c.Client.NewSearchBuilder(ctx, index, query)
}

func (c *OriginCmds) NewAggregateBuilder(ctx context.Context, index, query string) *redis.AggregateBuilder {
	return c.Client.NewAggregateBuilder(ctx, index, query)
}

func (c *OriginCmds) NewCreateIndexBuilder(ctx context.Context, index string) *redis.CreateIndexBuilder {
	return c.Client.NewCreateIndexBuilder(ctx, index)
}

func (c *OriginCmds) NewDropIndexBuilder(ctx context.Context, index string) *redis.DropIndexBuilder {
	return c.Client.NewDropIndexBuilder(ctx, index)
}

// RedisClient 返回底层原始的Client，用于需要直接访问底层 API 的特殊场景
func (c *OriginCmds) RedisClient() *redis.Client {
	if c == nil || c.Client == nil {
		return nil
	}
	return c.Client
}

// Eval 允许执行 Lua 脚本（EVAL）——传入的 keys 会被包装器加上前缀。
func (c *OriginCmds) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	if len(keys) == 0 {
		return c.Client.Eval(ctx, script, nil, args...)
	}
	return c.Client.Eval(ctx, script, keys, args...)
}
