package captchabase64

import (
	"context"
	"time"

	"github.com/ishaqcherry9/depend/pkg/goredis"
	"github.com/mojocn/base64Captcha"
)

// redisStore 实现 base64Captcha.Store 接口，使用 Redis 作为存储。
type redisStore struct {
	rdb        *goredis.Client
	keyPrefix  string
	expiration time.Duration
}

func newRedisStore(client *goredis.Client, keyPrefix string, expiration time.Duration) base64Captcha.Store {
	if keyPrefix == "" {
		keyPrefix = "captcha:"
	}
	if expiration <= 0 {
		expiration = 10 * time.Minute
	}
	return &redisStore{rdb: client, keyPrefix: keyPrefix, expiration: expiration}
}

func (s *redisStore) Set(id string, value string) error {
	if s == nil || s.rdb == nil || id == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.rdb.Set(ctx, s.keyPrefix+id, value, s.expiration).Err()
}

func (s *redisStore) Get(id string, clear bool) string {
	if s == nil || s.rdb == nil || id == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	key := s.keyPrefix + id
	val, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return ""
	}
	if clear {
		_ = s.rdb.Del(ctx, key).Err()
	}
	return val
}

func (s *redisStore) Verify(id, answer string, clear bool) bool {
	return s.Get(id, clear) == answer
}
