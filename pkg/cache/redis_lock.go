package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisLock 封装 Redis 分布式锁
type RedisLock struct {
	client     *redis.Client
	key        string        // 锁的 key
	value      string        // 锁的 value (唯一标识)
	duration   time.Duration // 锁的过期时间
	retryCount int           // 尝试获取锁的重试次数
	retryDelay time.Duration // 尝试获取锁的重试间隔
}

// NewRedisLock 创建一个新的 RedisLock 实例
// client: Redis 客户端
// key: 锁的 key
// duration: 锁的过期时间
// retryCount: 尝试获取锁的重试次数
// retryDelay: 尝试获取锁的重试间隔
func NewRedisLock(client *redis.Client, key string, duration time.Duration, retryCount int, retryDelay time.Duration) *RedisLock {
	return &RedisLock{
		client:     client,
		key:        key,
		value:      uuid.New().String(), // 使用 UUID 作为 value
		duration:   duration,
		retryCount: retryCount,
		retryDelay: retryDelay,
	}
}

// AcquireLock 获取锁 (加入重试机制)
func (l *RedisLock) AcquireLock(ctx context.Context) (bool, error) {
	for i := 0; i <= l.retryCount; i++ {
		// 使用 SETNX 命令 (SET if Not exists)
		ok, err := l.client.SetNX(ctx, l.key, l.value, l.duration).Result()
		if err != nil {
			return false, fmt.Errorf("获取锁失败: %w", err)
		}
		if ok {
			return true, nil
		}
		if i < l.retryCount {
			time.Sleep(l.retryDelay)
		}
	}
	return false, nil // 超过重试次数仍然失败
}

// ReleaseLock 释放锁 (使用 Lua 脚本保证原子性)
func (l *RedisLock) ReleaseLock(ctx context.Context) (bool, error) {
	// 使用 Lua 脚本保证原子性
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
	if err != nil {
		return false, fmt.Errorf("释放锁失败: %w", err)
	}

	// 脚本执行结果 1: 释放锁成功 0: 锁已被其他客户端持有
	if val, ok := result.(int64); ok && val == 1 {
		return true, nil
	}
	return false, nil
}

// Refresh 续约锁的过期时间 (优化点3: 自动续约)
func (l *RedisLock) Refresh(ctx context.Context) (bool, error) {
	// 检查锁是否存在，并更新过期时间
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value, l.duration.Milliseconds()).Result()
	if err != nil {
		return false, fmt.Errorf("续约锁失败: %w", err)
	}

	// 脚本执行结果 1: 续约锁成功 0: 锁已被其他客户端持有
	if val, ok := result.(int64); ok && val == 1 {
		return true, nil
	}

	return false, nil
}
