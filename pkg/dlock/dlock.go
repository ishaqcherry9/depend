package dlock

import "context"

type Locker interface {
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
	TryLock(ctx context.Context) (bool, error)
	Close() error
}
