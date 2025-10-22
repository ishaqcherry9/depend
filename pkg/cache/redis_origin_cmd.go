package cache

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"time"

	redisCmd "github.com/ishaqcherry9/depend/pkg/cache/redis"
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

var _ RedisOriginCmds = (*redisCmd.OriginCmds)(nil)

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
		RedisOriginCmds: &redisCmd.OriginCmds{Client: client},
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
		return fmt.Errorf("c.RedisOriginCmds.Set error: %v, cacheKey=%s", err, cacheKey)
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
		return fmt.Errorf("c.RedisOriginCmds.Del error: %v, keys=%+v", err, cacheKeys)
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
	pipeline := c.RedisOriginCmds.Pipeline()
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
		return fmt.Errorf("c.RedisOriginCmds.MGet error: %v, keys=%+v", err, cacheKeys)
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

	return c.RedisOriginCmds.Set(ctx, cacheKey, NotFoundPlaceholder, DefaultNotFoundExpireTime).Err()
}

// OriginSet 原生未做任何封装的Set命令
func (c *redisOriginCmdCache) OriginSet(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return c.RedisOriginCmds.Set(ctx, key, value, expiration)
}

// OriginGet 原生未做任何封装的Get命令
func (c *redisOriginCmdCache) OriginGet(ctx context.Context, key string) *redis.StringCmd {
	return c.RedisOriginCmds.Get(ctx, key)
}

// OriginDel 原生未做任何封装的Del命令
func (c *redisOriginCmdCache) OriginDel(ctx context.Context, keys ...string) *redis.IntCmd {
	if len(keys) == 0 {
		return c.RedisOriginCmds.Del(ctx)
	}
	return c.RedisOriginCmds.Del(ctx, keys...)
}

func (c *redisOriginCmdCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return c.RedisOriginCmds.SetNX(ctx, key, value, expiration)
}

func (c *redisOriginCmdCache) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	if len(keys) == 0 {
		return c.RedisOriginCmds.Exists(ctx)
	}
	return c.RedisOriginCmds.Exists(ctx, keys...)
}

func (c *redisOriginCmdCache) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return c.RedisOriginCmds.Expire(ctx, key, expiration)
}

func (c *redisOriginCmdCache) Incr(ctx context.Context, key string) *redis.IntCmd {
	return c.RedisOriginCmds.Incr(ctx, key)
}

func (c *redisOriginCmdCache) Decr(ctx context.Context, key string) *redis.IntCmd {
	return c.RedisOriginCmds.Decr(ctx, key)
}

func (c *redisOriginCmdCache) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.RedisOriginCmds.HSet(ctx, key, values...)
}

func (c *redisOriginCmdCache) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	return c.RedisOriginCmds.HGet(ctx, key, field)
}

func (c *redisOriginCmdCache) HMSet(ctx context.Context, key string, values ...interface{}) *redis.BoolCmd {
	return c.RedisOriginCmds.HMSet(ctx, key, values...)
}

func (c *redisOriginCmdCache) HMGet(ctx context.Context, key string, field ...string) *redis.SliceCmd {
	return c.RedisOriginCmds.HMGet(ctx, key, field...)
}

func (c *redisOriginCmdCache) HDel(ctx context.Context, key string, fields ...string) *redis.IntCmd {
	if len(fields) == 0 {
		return c.RedisOriginCmds.HDel(ctx, key)
	}
	return c.RedisOriginCmds.HDel(ctx, key, fields...)
}

func (c *redisOriginCmdCache) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.RedisOriginCmds.LPush(ctx, key, values...)
}

func (c *redisOriginCmdCache) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.RedisOriginCmds.RPush(ctx, key, values...)
}

func (c *redisOriginCmdCache) LPop(ctx context.Context, key string) *redis.StringCmd {
	return c.RedisOriginCmds.LPop(ctx, key)
}

func (c *redisOriginCmdCache) RPop(ctx context.Context, key string) *redis.StringCmd {
	return c.RedisOriginCmds.RPop(ctx, key)
}

func (c *redisOriginCmdCache) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return c.RedisOriginCmds.SAdd(ctx, key, members...)
}

func (c *redisOriginCmdCache) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	return c.RedisOriginCmds.SMembers(ctx, key)
}

func (c *redisOriginCmdCache) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	return c.RedisOriginCmds.ZAdd(ctx, key, members...)
}

func (c *redisOriginCmdCache) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	return c.RedisOriginCmds.ZRange(ctx, key, start, stop)
}

func (c *redisOriginCmdCache) ZRangeWithScores(ctx context.Context, key string, start, stop int64) *redis.ZSliceCmd {
	return c.RedisOriginCmds.ZRangeWithScores(ctx, key, start, stop)
}

func (c *redisOriginCmdCache) ZIncrBy(ctx context.Context, key string, increment float64, member string) *redis.FloatCmd {
	return c.RedisOriginCmds.ZIncrBy(ctx, key, increment, member)
}

func (c *redisOriginCmdCache) ZRem(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	return c.RedisOriginCmds.ZRem(ctx, key, members...)
}

func (c *redisOriginCmdCache) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	return c.RedisOriginCmds.Publish(ctx, channel, message)
}

func (c *redisOriginCmdCache) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.RedisOriginCmds.Subscribe(ctx)
	}
	return c.RedisOriginCmds.Subscribe(ctx, channels...)
}

func (c *redisOriginCmdCache) SetArgs(ctx context.Context, key string, value interface{}, a redis.SetArgs) *redis.StatusCmd {
	return c.RedisOriginCmds.SetArgs(ctx, key, value, a)
}

func (c *redisOriginCmdCache) Pipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return c.RedisOriginCmds.Pipelined(ctx, fn)
}

func (c *redisOriginCmdCache) Pipeline() redis.Pipeliner {
	return c.RedisOriginCmds.Pipeline()
}

func (c *redisOriginCmdCache) TxPipelined(ctx context.Context, fn func(redis.Pipeliner) error) ([]redis.Cmder, error) {
	return c.RedisOriginCmds.TxPipelined(ctx, fn)
}

func (c *redisOriginCmdCache) TxPipeline() redis.Pipeliner {
	return c.RedisOriginCmds.TxPipeline()
}

func (c *redisOriginCmdCache) Process(ctx context.Context, cmd redis.Cmder) error {
	return c.RedisOriginCmds.Process(ctx, cmd)
}

func (c *redisOriginCmdCache) Options() *redis.Options {
	return c.RedisOriginCmds.Options()
}

func (c *redisOriginCmdCache) PoolStats() *redis.PoolStats {
	return c.RedisOriginCmds.PoolStats()
}

func (c *redisOriginCmdCache) PSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.RedisOriginCmds.PSubscribe(ctx)
	}
	return c.RedisOriginCmds.PSubscribe(ctx, channels...)
}

func (c *redisOriginCmdCache) SSubscribe(ctx context.Context, channels ...string) *redis.PubSub {
	if len(channels) == 0 {
		return c.RedisOriginCmds.SSubscribe(ctx)
	}
	return c.RedisOriginCmds.SSubscribe(ctx, channels...)
}

func (c *redisOriginCmdCache) NewSearchBuilder(ctx context.Context, index, query string) *redis.SearchBuilder {
	return c.RedisOriginCmds.NewSearchBuilder(ctx, index, query)
}

func (c *redisOriginCmdCache) NewAggregateBuilder(ctx context.Context, index, query string) *redis.AggregateBuilder {
	return c.RedisOriginCmds.NewAggregateBuilder(ctx, index, query)
}

func (c *redisOriginCmdCache) NewCreateIndexBuilder(ctx context.Context, index string) *redis.CreateIndexBuilder {
	return c.RedisOriginCmds.NewCreateIndexBuilder(ctx, index)
}

func (c *redisOriginCmdCache) NewDropIndexBuilder(ctx context.Context, index string) *redis.DropIndexBuilder {
	return c.RedisOriginCmds.NewDropIndexBuilder(ctx, index)
}

// RedisClient 返回底层原始的Client，用于需要直接访问底层 API 的特殊场景
func (c *redisOriginCmdCache) RedisClient() *redis.Client {
	if c == nil || c.RedisOriginCmds == nil {
		return nil
	}
	return c.client
}

// Eval 允许执行 Lua 脚本（EVAL）——传入的 keys 会被包装器加上前缀。
func (c *redisOriginCmdCache) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	if len(keys) == 0 {
		return c.RedisOriginCmds.Eval(ctx, script, nil, args...)
	}
	return c.RedisOriginCmds.Eval(ctx, script, keys, args...)
}
