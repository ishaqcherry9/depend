package cache

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/ishaqcherry9/depend/pkg/encoding"
)

var CacheNotFound = redis.Nil

type redisCache struct {
	client            *redis.Client
	KeyPrefix         string
	encoding          encoding.Encoding
	DefaultExpireTime time.Duration
	newObject         func() interface{}
}

func NewRedisCache(client *redis.Client, keyPrefix string, encode encoding.Encoding, newObject func() interface{}) Cache {
	return &redisCache{
		client:    client,
		KeyPrefix: keyPrefix,
		encoding:  encode,
		newObject: newObject,
	}
}

func (c *redisCache) Set(ctx context.Context, key string, val interface{}, expiration time.Duration) error {
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

func (c *redisCache) Get(ctx context.Context, key string, val interface{}) error {
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

func (c *redisCache) MultiSet(ctx context.Context, valueMap map[string]interface{}, expiration time.Duration) error {
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

func (c *redisCache) MultiGet(ctx context.Context, keys []string, value interface{}) error {
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

func (c *redisCache) Del(ctx context.Context, keys ...string) error {
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

func (c *redisCache) SetCacheWithNotFound(ctx context.Context, key string) error {
	cacheKey, err := BuildCacheKey(c.KeyPrefix, key)
	if err != nil {
		return fmt.Errorf("BuildCacheKey error: %v, key=%s", err, key)
	}

	return c.client.Set(ctx, cacheKey, NotFoundPlaceholder, DefaultNotFoundExpireTime).Err()
}

func BuildCacheKey(keyPrefix string, key string) (string, error) {
	if key == "" {
		return "", errors.New("[cache] key should not be empty")
	}

	cacheKey := key
	if keyPrefix != "" {
		cacheKey = strings.Join([]string{keyPrefix, key}, ":")
	}

	return cacheKey, nil
}

type redisClusterCache struct {
	client            *redis.ClusterClient
	KeyPrefix         string
	encoding          encoding.Encoding
	DefaultExpireTime time.Duration
	newObject         func() interface{}
}

func NewRedisClusterCache(client *redis.ClusterClient, keyPrefix string, encode encoding.Encoding, newObject func() interface{}) Cache {
	return &redisClusterCache{
		client:    client,
		KeyPrefix: keyPrefix,
		encoding:  encode,
		newObject: newObject,
	}
}

func (c *redisClusterCache) Set(ctx context.Context, key string, val interface{}, expiration time.Duration) error {
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

func (c *redisClusterCache) Get(ctx context.Context, key string, val interface{}) error {
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

func (c *redisClusterCache) MultiSet(ctx context.Context, valueMap map[string]interface{}, expiration time.Duration) error {
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

func (c *redisClusterCache) MultiGet(ctx context.Context, keys []string, value interface{}) error {
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
			fmt.Printf("unmarshal data error: %+v, cacheKey=%s type=%T\n", err, cacheKeys[i], value)
			continue
		}
		valueMap.SetMapIndex(reflect.ValueOf(cacheKeys[i]), reflect.ValueOf(object))
	}
	return nil
}

func (c *redisClusterCache) Del(ctx context.Context, keys ...string) error {
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

func (c *redisClusterCache) SetCacheWithNotFound(ctx context.Context, key string) error {
	cacheKey, err := BuildCacheKey(c.KeyPrefix, key)
	if err != nil {
		return fmt.Errorf("BuildCacheKey error: %v, key=%s", err, key)
	}

	return c.client.Set(ctx, cacheKey, NotFoundPlaceholder, DefaultNotFoundExpireTime).Err()
}
