package dlock

import (
	"context"
	"time"
)

/*
Locker 接口定义了锁的基本操作

	1.SpinLock (自旋锁)
	2.AtomicMutex (基于原子操作的互斥锁)
	3.CASLock (基于 CAS 的锁)
	4.RWLock (读写锁)
	5.TicketLock (票据锁)
	6.ALock (数组锁)
	7.CLHLock (CLH 锁)
	8.RedisLock (基于 Redis 的分布式锁)
*/
type Locker interface {
	Lock(ctx context.Context) error                                   // 获取锁，阻塞直到获取成功或发生错误
	Unlock(ctx context.Context) error                                 // 释放锁
	TryLock(ctx context.Context, timeout time.Duration) (bool, error) // 尝试在指定时间内获取锁
	Close() error                                                     // 释放锁相关资源，例如关闭连接，停止 goroutine
}
