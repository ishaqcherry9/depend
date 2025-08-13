package cache

import (
	"context"
	"errors"
	"time"
)

var (
	DefaultExpireTime         = time.Hour * 24
	DefaultNotFoundExpireTime = time.Minute * 10

	NotFoundPlaceholder      = "*"
	NotFoundPlaceholderBytes = []byte(NotFoundPlaceholder)
	ErrPlaceholder           = errors.New("cache: placeholder")

	DefaultClient Cache
)

type Cache interface {
	Set(ctx context.Context, key string, val interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, val interface{}) error
	MultiSet(ctx context.Context, valMap map[string]interface{}, expiration time.Duration) error
	MultiGet(ctx context.Context, keys []string, valueMap interface{}) error
	Del(ctx context.Context, keys ...string) error
	SetCacheWithNotFound(ctx context.Context, key string) error
}

func Set(ctx context.Context, key string, val interface{}, expiration time.Duration) error {
	return DefaultClient.Set(ctx, key, val, expiration)
}

func Get(ctx context.Context, key string, val interface{}) error {
	return DefaultClient.Get(ctx, key, val)
}

func MultiSet(ctx context.Context, valMap map[string]interface{}, expiration time.Duration) error {
	return DefaultClient.MultiSet(ctx, valMap, expiration)
}

func MultiGet(ctx context.Context, keys []string, valueMap interface{}) error {
	return DefaultClient.MultiGet(ctx, keys, valueMap)
}

func Del(ctx context.Context, keys ...string) error {
	return DefaultClient.Del(ctx, keys...)
}

func SetCacheWithNotFound(ctx context.Context, key string) error {
	return DefaultClient.SetCacheWithNotFound(ctx, key)
}
