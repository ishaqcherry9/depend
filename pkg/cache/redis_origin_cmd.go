package cache

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
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

type redisOriginCmds struct {
	client *redis.Client
}

var _ RedisOriginCmds = (*redisOriginCmds)(nil)

type redisOriginCmdCache struct {
	client            *redis.Client
	KeyPrefix         string
	encoding          encoding.Encoding
	DefaultExpireTime time.Duration
	newObject         func() interface{}
	RedisOriginCmds
}

func NewRedisOriginCmdCache(client *redis.Client, keyPrefix string, encode encoding.Encoding, newObject func() interface{}) Cache {
	return &redisOriginCmdCache{
		client:          client,
		KeyPrefix:       keyPrefix,
		encoding:        encode,
		newObject:       newObject,
		RedisOriginCmds: &redisOriginCmds{client: client},
	}
}

func (c *redisOriginCmdCache) Set(ctx context.Context, key string, val interface{}, expiration time.Duration) error {
	buf, err := encoding.Marshal(c.encoding, val)
	if err != nil {
		return fmt.Errorf("encoding.Marshal error: %v, key=%s, val=%+v ", err, key, val)
	}

	cacheKey, err := BuildCacheKey(c.KeyPrefix, key)
	if err != nil {
		return fmt.Errorf("BuildCacheKey error: %v, key=%s", err, key)
	}

	if len(buf) == 0 {
		buf = NotFoundPlaceholderBytes
	}
	err = c.client.Set(ctx, cacheKey, buf, expiration).Err()
	if err != nil {
		return fmt.Errorf("c.client.Set error: %v, cacheKey=%s", err, cacheKey)
	}
	return nil
}

func (c *redisOriginCmdCache) Get(ctx context.Context, key string, val interface{}) error {
	cacheKey, err := BuildCacheKey(c.KeyPrefix, key)
	if err != nil {
		return fmt.Errorf("BuildCacheKey error: %v, key=%s", err, key)
	}

	dataBytes, err := c.client.Get(ctx, cacheKey).Bytes()
	if err != nil {
		return err
	}

	if len(dataBytes) == 0 || bytes.Equal(dataBytes, NotFoundPlaceholderBytes) {
		return ErrPlaceholder
	}
	err = encoding.Unmarshal(c.encoding, dataBytes, val)
	if err != nil {
		return fmt.Errorf("encoding.Unmarshal error: %v, key=%s, cacheKey=%s, type=%T, json=%s ",
			err, key, cacheKey, val, dataBytes)
	}
	return nil
}

func (c *redisOriginCmdCache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	cacheKeys := make([]string, len(keys))
	for index, key := range keys {
		cacheKey, err := BuildCacheKey(c.KeyPrefix, key)
		if err != nil {
			continue
		}
		cacheKeys[index] = cacheKey
	}
	err := c.client.Del(ctx, cacheKeys...).Err()
	if err != nil {
		return fmt.Errorf("c.client.Del error: %v, keys=%+v", err, cacheKeys)
	}
	return nil
}

func (c *redisOriginCmdCache) MultiSet(ctx context.Context, valueMap map[string]interface{}, expiration time.Duration) error {
	if len(valueMap) == 0 {
		return nil
	}

	paris := make([]interface{}, 0, 2*len(valueMap))
	for key, value := range valueMap {
		buf, err := encoding.Marshal(c.encoding, value)
		if err != nil {
			fmt.Printf("encoding.Marshal error, %v, value:%v\n", err, value)
			continue
		}
		cacheKey, err := BuildCacheKey(c.KeyPrefix, key)
		if err != nil {
			fmt.Printf("BuildCacheKey error, %v, key:%v\n", err, key)
			continue
		}
		paris = append(paris, []byte(cacheKey))
		paris = append(paris, buf)
	}
	pipeline := c.client.Pipeline()
	err := pipeline.MSet(ctx, paris...).Err()
	if err != nil {
		return fmt.Errorf("pipeline.MSet error: %v", err)
	}
	for i := 0; i < len(paris); i = i + 2 {
		switch paris[i].(type) {
		case []byte:
			pipeline.Expire(ctx, string(paris[i].([]byte)), expiration)
		default:
			fmt.Printf("redis expire is unsupported key type: %T\n", paris[i])
		}
	}
	_, err = pipeline.Exec(ctx)
	if err != nil {
		return fmt.Errorf("pipeline.Exec error: %v", err)
	}
	return nil
}

func (c *redisOriginCmdCache) MultiGet(ctx context.Context, keys []string, value interface{}) error {
	if len(keys) == 0 {
		return nil
	}
	cacheKeys := make([]string, len(keys))
	for index, key := range keys {
		cacheKey, err := BuildCacheKey(c.KeyPrefix, key)
		if err != nil {
			return fmt.Errorf("BuildCacheKey error: %v, key=%s", err, key)
		}
		cacheKeys[index] = cacheKey
	}
	values, err := c.client.MGet(ctx, cacheKeys...).Result()
	if err != nil {
		return fmt.Errorf("c.client.MGet error: %v, keys=%+v", err, cacheKeys)
	}

	valueMap := reflect.ValueOf(value)
	for i, v := range values {
		if v == nil {
			continue
		}
		dataBytes := []byte(v.(string))
		if len(dataBytes) == 0 || bytes.Equal(dataBytes, NotFoundPlaceholderBytes) {
			continue
		}
		object := c.newObject()
		err = encoding.Unmarshal(c.encoding, dataBytes, object)
		if err != nil {
			fmt.Printf("unmarshal data error: %+v, cacheKey=%s valueType=%T\n", err, cacheKeys[i], value)
			continue
		}
		valueMap.SetMapIndex(reflect.ValueOf(cacheKeys[i]), reflect.ValueOf(object))
	}
	return nil
}

func (c *redisOriginCmdCache) SetCacheWithNotFound(ctx context.Context, key string) error {
	cacheKey, err := BuildCacheKey(c.KeyPrefix, key)
	if err != nil {
		return fmt.Errorf("BuildCacheKey error: %v, key=%s", err, key)
	}

	return c.client.Set(ctx, cacheKey, NotFoundPlaceholder, DefaultNotFoundExpireTime).Err()
}

func (c *redisOriginCmds) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return c.client.Set(ctx, key, value, expiration)
}

func (c *redisOriginCmds) Get(ctx context.Context, key string) *redis.StringCmd {
	return c.client.Get(ctx, key)
}

func (c *redisOriginCmds) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	if len(keys) == 0 {
		return c.client.Del(ctx)
	}
	return c.client.Del(ctx, keys...)
}

func (c *redisOriginCmds) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return c.client.SetNX(ctx, key, value, expiration)
}

func (c *redisOriginCmds) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	if len(keys) == 0 {
		return c.client.Exists(ctx)
	}
	return c.client.Exists(ctx, keys...)
}

func (c *redisOriginCmds) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return c.client.Expire(ctx, key, expiration)
}

func (c *redisOriginCmds) Incr(ctx context.Context, key string) *redis.IntCmd {
	return c.client.Incr(ctx, key)
}

func (c *redisOriginCmds) Decr(ctx context.Context, key string) *redis.IntCmd {
	return c.client.Decr(ctx, key)
}

func (c *redisOriginCmds) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.client.HSet(ctx, key, values...)
}

func (c *redisOriginCmds) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return c.client.HGet(ctx, key, field)
}

func (c *redisOriginCmds) HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd {
	return c.client.HMSet(ctx, key, values...)
}

func (c *redisOriginCmds) HMGet(ctx context.Context, key string, field ...string) *redis.SliceCmd {
	return c.client.HMGet(ctx, key, field...)
}

func (c *redisOriginCmds) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	if len(fields) == 0 {
		return c.client.HDel(ctx, key)
	}
	return c.client.HDel(ctx, key, fields...)
}

func (c *redisOriginCmds) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.client.LPush(ctx, key, values...)
}

func (c *redisOriginCmds) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.client.RPush(ctx, key, values...)
}

func (c *redisOriginCmds) LPop(ctx context.Context, key string) *redis.StringCmd {
	return c.client.LPop(ctx, key)
}

func (c *redisOriginCmds) RPop(ctx context.Context, key string) *redis.StringCmd {
	return c.client.RPop(ctx, key)
}

func (c *redisOriginCmds) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return c.client.SAdd(ctx, key, members...)
}

func (c *redisOriginCmds) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	return c.client.SMembers(ctx, key)
}

func (c *redisOriginCmds) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	return c.client.ZAdd(ctx, key, members...)
}

func (c *redisOriginCmds) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return c.client.ZRange(ctx, key, start, stop)
}

func (c *redisOriginCmds) ZRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	return c.client.ZRangeWithScores(ctx, key, start, stop)
}

func (c *redisOriginCmds) ZIncrBy(ctx context.Context, key string, increment float64, member string) *redis.FloatCmd {
	return c.client.ZIncrBy(ctx, key, increment, member)
}

func (c *redisOriginCmds) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return c.client.ZRem(ctx, key, members...)
}

func (c *redisOriginCmds) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	return c.client.Publish(ctx, channel, message)
}

func (c *redisOriginCmds) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.client.Subscribe(ctx)
	}
	return c.client.Subscribe(ctx, channels...)
}

func (c *redisOriginCmds) SetArgs(ctx context.Context, key string, value interface{}, a redis.SetArgs) *redis.StatusCmd {
	return c.client.SetArgs(ctx, key, value, a)
}

func (c *redisOriginCmds) Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return c.client.Pipelined(ctx, fn)
}

func (c *redisOriginCmds) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

func (c *redisOriginCmds) TxPipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return c.client.TxPipelined(ctx, fn)
}

func (c *redisOriginCmds) TxPipeline() redis.Pipeliner {
	return c.client.TxPipeline()
}

func (c *redisOriginCmds) Process(ctx context.Context, cmd redis.Cmder) error {
	return c.client.Process(ctx, cmd)
}

func (c *redisOriginCmds) Options() *redis.Options {
	return c.client.Options()
}

func (c *redisOriginCmds) PoolStats() *redis.PoolStats {
	return c.client.PoolStats()
}

func (c *redisOriginCmds) PSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.client.PSubscribe(ctx)
	}
	return c.client.PSubscribe(ctx, channels...)
}

func (c *redisOriginCmds) SSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.client.SSubscribe(ctx)
	}
	return c.client.SSubscribe(ctx, channels...)
}

func (c *redisOriginCmds) NewSearchBuilder(ctx context.Context, index, query string) *redis.SearchBuilder {
	return c.client.NewSearchBuilder(ctx, index, query)
}

func (c *redisOriginCmds) NewAggregateBuilder(ctx context.Context, index, query string) *redis.AggregateBuilder {
	return c.client.NewAggregateBuilder(ctx, index, query)
}

func (c *redisOriginCmds) NewCreateIndexBuilder(ctx context.Context, index string) *redis.CreateIndexBuilder {
	return c.client.NewCreateIndexBuilder(ctx, index)
}

func (c *redisOriginCmds) NewDropIndexBuilder(ctx context.Context, index string) *redis.DropIndexBuilder {
	return c.client.NewDropIndexBuilder(ctx, index)
}

// RedisClient 返回底层原始的Client，用于需要直接访问底层 API 的特殊场景
func (c *redisOriginCmds) RedisClient() *redis.Client {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client
}

// Eval 允许执行 Lua 脚本（EVAL）——传入的 keys 会被包装器加上前缀。
func (c *redisOriginCmds) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	if len(keys) == 0 {
		return c.client.Eval(ctx, script, nil, args...)
	}
	return c.client.Eval(ctx, script, keys, args...)
}
