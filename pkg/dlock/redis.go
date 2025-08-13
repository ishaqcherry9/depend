package dlock

import (
	"context"
	"errors"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

type RedisLock struct {
	mutex *redsync.Mutex
}

func NewRedisLock(client *redis.Client, key string, options ...redsync.Option) (Locker, error) {
	if client == nil {
		return nil, errors.New("redis client is nil")
	}
	if key == "" {
		return nil, errors.New("key is empty")
	}
	return newLocker(client, key, options...), nil
}

func NewRedisClusterLock(clusterClient *redis.ClusterClient, key string, options ...redsync.Option) (Locker, error) {
	if clusterClient == nil {
		return nil, errors.New("cluster redis client is nil")
	}
	if key == "" {
		return nil, errors.New("key is empty")
	}
	return newLocker(clusterClient, key, options...), nil
}

func newLocker(delegate redis.UniversalClient, key string, options ...redsync.Option) Locker {
	pool := goredis.NewPool(delegate)
	rs := redsync.New(pool)
	mutex := rs.NewMutex(key, options...)

	return &RedisLock{
		mutex: mutex,
	}
}

func (l *RedisLock) TryLock(ctx context.Context) (bool, error) {
	err := l.mutex.TryLockContext(ctx)
	if err == nil {
		return true, nil
	}
	return false, err
}

func (l *RedisLock) Lock(ctx context.Context) error {
	return l.mutex.LockContext(ctx)
}

func (l *RedisLock) Unlock(ctx context.Context) error {
	_, err := l.mutex.UnlockContext(ctx)
	return err
}

func (l *RedisLock) Close() error {
	return nil
}
