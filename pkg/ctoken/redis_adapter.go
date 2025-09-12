package ctoken

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/ishaqcherry9/depend/pkg/goredis"
	"time"
)

type Func = func(ctx context.Context) (value interface{}, err error)

type RedisAdapter struct {
	redis *goredis.Client
}

func NewRedisAdapter(redis *goredis.Client) *RedisAdapter {
	return &RedisAdapter{
		redis: redis,
	}
}

func (c *RedisAdapter) Set(ctx context.Context, key interface{}, value interface{}, duration time.Duration) (err error) {
	redisKey := gconv.String(key)
	if value == nil || duration < 0 {
		err = c.redis.Del(ctx, redisKey).Err()

	} else {
		if duration == 0 {
			err = c.redis.Set(ctx, redisKey, value, 0).Err()
		} else {
			err = c.redis.Set(ctx, redisKey, value, duration).Err()
		}
	}
	return err
}

func (c *RedisAdapter) Get(ctx context.Context, key interface{}) (*gvar.Var, error) {
	redisKey := gconv.String(key)
	val, err := c.redis.Get(ctx, redisKey).Result()
	if err != nil {
		if errors.Is(err, goredis.ErrRedisNotFound) {
			return gvar.New(nil), nil
		}
		return nil, err
	}
	return gvar.New(val), nil
}

func (c *RedisAdapter) Remove(ctx context.Context, keys ...interface{}) (lastValue *gvar.Var, err error) {
	if len(keys) == 0 {
		return nil, nil
	}
	val, err := c.redis.Get(ctx, gconv.String(keys[len(keys)-1])).Result()
	if err != nil {
		return nil, err
	}
	lastValue = gvar.New(val)
	err = c.redis.Del(ctx, gconv.Strings(keys)...).Err()
	return
}

func (c *RedisAdapter) GetExpire(ctx context.Context, key interface{}) (time.Duration, error) {
	return 0, nil
}

func (c *RedisAdapter) SetMap(ctx context.Context, data map[interface{}]interface{}, duration time.Duration) error {
	return nil
}

func (c *RedisAdapter) SetIfNotExist(ctx context.Context, key interface{}, value interface{}, duration time.Duration) (bool, error) {
	return false, nil
}

func (c *RedisAdapter) SetIfNotExistFunc(ctx context.Context, key interface{}, f Func, duration time.Duration) (ok bool, err error) {
	return false, nil
}

func (c *RedisAdapter) SetIfNotExistFuncLock(ctx context.Context, key interface{}, f Func, duration time.Duration) (ok bool, err error) {
	return false, nil
}

func (c *RedisAdapter) GetOrSet(ctx context.Context, key interface{}, value interface{}, duration time.Duration) (result *gvar.Var, err error) {

	return
}

func (c *RedisAdapter) GetOrSetFunc(ctx context.Context, key interface{}, f Func, duration time.Duration) (result *gvar.Var, err error) {
	return nil, nil

}

func (c *RedisAdapter) GetOrSetFuncLock(ctx context.Context, key interface{}, f Func, duration time.Duration) (result *gvar.Var, err error) {
	return nil, nil
}

// Contains checks and returns true if `key` exists in the cache, or else returns false.
func (c *RedisAdapter) Contains(ctx context.Context, key interface{}) (bool, error) {
	return false, nil
}

// Size returns the number of items in the cache.
func (c *RedisAdapter) Size(ctx context.Context) (size int, err error) {
	return 0, nil
}

func (c *RedisAdapter) Data(ctx context.Context) (map[interface{}]interface{}, error) {
	return nil, nil
}

func (c *RedisAdapter) Keys(ctx context.Context) ([]interface{}, error) {
	return nil, nil
}

func (c *RedisAdapter) Values(ctx context.Context) ([]interface{}, error) {

	return nil, nil
}

func (c *RedisAdapter) Update(ctx context.Context, key interface{}, value interface{}) (oldValue *gvar.Var, exist bool, err error) {

	return nil, false, nil
}

func (c *RedisAdapter) UpdateExpire(ctx context.Context, key interface{}, duration time.Duration) (oldDuration time.Duration, err error) {

	return
}

func (c *RedisAdapter) Clear(ctx context.Context) (err error) {
	return nil
}

// Close closes the cache.
func (c *RedisAdapter) Close(ctx context.Context) error {
	// It does nothing.
	return nil
}
