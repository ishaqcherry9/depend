package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/ishaqcherry9/depend/pkg/encoding"
	jetcache "github.com/mgtv-tech/jetcache-go"
	_ "github.com/mgtv-tech/jetcache-go/encoding/json"
	"github.com/mgtv-tech/jetcache-go/local"
	"github.com/mgtv-tech/jetcache-go/remote"
	"github.com/redis/go-redis/v9"
	"reflect"
	"time"
)

type tieredCache struct {
	jetCacheInstance jetcache.Cache
	KeyPrefix        string
	encoding         encoding.Encoding
	newObject        func() interface{}
}

type TieredConfig struct {
	LocalCacheSize int64
	LocalTTL       time.Duration
	RemoteTTL      time.Duration
	EnableStats    bool
}

func NewTieredCache(client *redis.Client, keyPrefix string, encode encoding.Encoding, newObject func() interface{}) Cache {
	return NewTieredCacheWithConfig(client, keyPrefix, encode, newObject, TieredConfig{})
}

func NewTieredCacheWithConfig(client *redis.Client, keyPrefix string, encode encoding.Encoding, newObject func() interface{}, config TieredConfig) Cache {
	if config.LocalCacheSize <= 0 {
		config.LocalCacheSize = 256
	}
	if config.LocalTTL <= 0 {
		config.LocalTTL = time.Minute * 5
	}
	if config.RemoteTTL <= 0 {
		config.RemoteTTL = time.Hour
	}

	opts := []jetcache.Option{
		jetcache.WithName(fmt.Sprintf("tiered_%s", keyPrefix)),
		jetcache.WithCodec("json"),
		jetcache.WithRemoteExpiry(config.RemoteTTL),
		jetcache.WithStatsDisabled(!config.EnableStats),
	}

	localCache := local.NewFreeCache(
		local.Size(config.LocalCacheSize)*local.MB,
		config.LocalTTL,
	)
	opts = append(opts, jetcache.WithLocal(localCache))

	if client != nil {
		remoteCache := remote.NewGoRedisV9Adapter(client)
		opts = append(opts, jetcache.WithRemote(remoteCache))
	}

	jetCacheInstance := jetcache.New(opts...)

	return &tieredCache{
		jetCacheInstance: jetCacheInstance,
		KeyPrefix:        keyPrefix,
		encoding:         encode,
		newObject:        newObject,
	}
}

func NewTieredClusterCache(client *redis.ClusterClient, keyPrefix string, encode encoding.Encoding, newObject func() interface{}) Cache {
	return NewTieredClusterCacheWithConfig(client, keyPrefix, encode, newObject, TieredConfig{})
}

func NewTieredClusterCacheWithConfig(client *redis.ClusterClient, keyPrefix string, encode encoding.Encoding, newObject func() interface{}, config TieredConfig) Cache {

	if config.LocalCacheSize <= 0 {
		config.LocalCacheSize = 256
	}
	if config.LocalTTL <= 0 {
		config.LocalTTL = time.Minute * 5
	}
	if config.RemoteTTL <= 0 {
		config.RemoteTTL = time.Hour
	}

	opts := []jetcache.Option{
		jetcache.WithName(fmt.Sprintf("tiered_cluster_%s", keyPrefix)),
		jetcache.WithCodec("json"),
		jetcache.WithRemoteExpiry(config.RemoteTTL),
		jetcache.WithStatsDisabled(!config.EnableStats),
	}

	localCache := local.NewFreeCache(
		local.Size(config.LocalCacheSize)*local.MB,
		config.LocalTTL,
	)
	opts = append(opts, jetcache.WithLocal(localCache))

	if client != nil {
		remoteCache := remote.NewGoRedisV9Adapter(client)
		opts = append(opts, jetcache.WithRemote(remoteCache))
	}

	jetCacheInstance := jetcache.New(opts...)

	return &tieredCache{
		jetCacheInstance: jetCacheInstance,
		KeyPrefix:        keyPrefix,
		encoding:         encode,
		newObject:        newObject,
	}
}

func (t *tieredCache) Set(ctx context.Context, key string, val interface{}, expiration time.Duration) error {
	cacheKey, err := BuildCacheKey(t.KeyPrefix, key)
	if err != nil {
		return fmt.Errorf("BuildCacheKey error: %v, key=%s", err, key)
	}

	return t.jetCacheInstance.Set(ctx, cacheKey, jetcache.Value(val), jetcache.TTL(expiration))
}

func (t *tieredCache) Get(ctx context.Context, key string, val interface{}) error {
	cacheKey, err := BuildCacheKey(t.KeyPrefix, key)
	if err != nil {
		return fmt.Errorf("BuildCacheKey error: %v, key=%s", err, key)
	}

	err = t.jetCacheInstance.Get(ctx, cacheKey, val)
	if err != nil {
		if errors.Is(err, jetcache.ErrCacheMiss) {
			return CacheNotFound
		}
		return err
	}
	return nil
}

func (t *tieredCache) MultiSet(ctx context.Context, valMap map[string]interface{}, expiration time.Duration) error {
	for key, val := range valMap {
		if err := t.Set(ctx, key, val, expiration); err != nil {
			return err
		}
	}
	return nil
}

func (t *tieredCache) MultiGet(ctx context.Context, keys []string, value interface{}) error {
	if len(keys) == 0 {
		return nil
	}

	valueMap := reflect.ValueOf(value)
	if valueMap.Kind() != reflect.Ptr || valueMap.Elem().Kind() != reflect.Map {
		return fmt.Errorf("value must be a pointer to map")
	}

	mapValue := valueMap.Elem()
	if mapValue.IsNil() {
		mapValue.Set(reflect.MakeMap(mapValue.Type()))
	}

	for _, key := range keys {
		if t.newObject == nil {
			continue
		}

		cacheKey, err := BuildCacheKey(t.KeyPrefix, key)
		if err != nil {
			continue
		}

		obj := t.newObject()
		err = t.Get(ctx, key, obj)
		if err != nil {
			continue
		}

		mapValue.SetMapIndex(reflect.ValueOf(cacheKey), reflect.ValueOf(obj))
	}

	return nil
}

func (t *tieredCache) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		cacheKey, err := BuildCacheKey(t.KeyPrefix, key)
		if err != nil {
			continue
		}

		if err := t.jetCacheInstance.Delete(ctx, cacheKey); err != nil {
			fmt.Printf("[WARN] tiered cache delete failed: key=%s, err=%v\n", cacheKey, err)
		}
	}
	return nil
}

func (t *tieredCache) SetCacheWithNotFound(ctx context.Context, key string) error {
	cacheKey, err := BuildCacheKey(t.KeyPrefix, key)
	if err != nil {
		return fmt.Errorf("BuildCacheKey error: %v, key=%s", err, key)
	}

	return t.jetCacheInstance.Set(ctx, cacheKey,
		jetcache.Value(NotFoundPlaceholder),
		jetcache.TTL(DefaultNotFoundExpireTime))
}

func (t *tieredCache) Close() {
	t.jetCacheInstance.Close()
}

func (t *tieredCache) GetCacheType() string {
	return t.jetCacheInstance.CacheType()
}

func (t *tieredCache) TaskSize() int {
	return t.jetCacheInstance.TaskSize()
}
