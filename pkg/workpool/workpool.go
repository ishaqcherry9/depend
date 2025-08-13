package workerpool

import (
	"github.com/ishaqcherry9/depend/pkg/logger"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
	"runtime/debug"
	"time"
)

const (
	antsPoolSize     = 1 << 2
	maxBlockingTasks = 500
	releaseTimeout   = 1 << 1
	preAlloc         = true
	nonblocking      = false
	expiryDuration   = 10 * time.Second
)

type (
	WorkerFIFOPool = ants.Pool
	WorkerLIFOPool = ants.Pool
)

func NewWorkerFIFOPool(size int, fn ...func(interface{})) *WorkerFIFOPool {
	opts := defaultWorkerPoolOpts()
	opts.PreAlloc = true
	if len(fn) > 0 {
		opts.PanicHandler = fn[0]
	}
	pool, err := ants.NewPool(size, ants.WithOptions(*opts))
	if err != nil {
		logger.Fatal("ants new pool failed", zap.Error(err))
	}
	return pool
}

func NewWorkerCustomFIFOPool(size int, expiryDuration int64, maxBlockingTasks int, nonblocking bool, fn ...func(interface{})) *WorkerFIFOPool {
	opts := defaultWorkerPoolOpts()
	opts.ExpiryDuration = time.Duration(expiryDuration) * time.Second
	opts.MaxBlockingTasks = maxBlockingTasks
	opts.PreAlloc = true
	opts.Nonblocking = nonblocking
	if len(fn) > 0 {
		opts.PanicHandler = fn[0]
	}
	pool, err := ants.NewPool(size, ants.WithOptions(*opts))
	if err != nil {
		logger.Fatal("ants new pool failed", zap.Error(err))
	}
	return pool
}

func NewWorkerLIFOPool(size int, fn ...func(interface{})) *WorkerLIFOPool {
	opts := defaultWorkerPoolOpts()
	opts.PreAlloc = false
	if len(fn) > 0 {
		opts.PanicHandler = fn[0]
	}
	pool, err := ants.NewPool(size, ants.WithOptions(*opts))
	if err != nil {
		logger.Fatal("ants new pool failed", zap.Error(err))
	}
	return pool
}

func NewWorkerCustomLIFOPool(size int, expiryDuration int64, maxBlockingTasks int, preAlloc bool, nonblocking bool, fn ...func(interface{})) *WorkerLIFOPool {
	opts := defaultWorkerPoolOpts()
	opts.ExpiryDuration = time.Duration(expiryDuration) * time.Second
	opts.MaxBlockingTasks = maxBlockingTasks
	opts.PreAlloc = false
	opts.Nonblocking = nonblocking
	if len(fn) > 0 {
		opts.PanicHandler = fn[0]
	}
	pool, err := ants.NewPool(size, ants.WithOptions(*opts))
	if err != nil {
		logger.Fatal("ants new pool failed", zap.Error(err))
	}
	return pool
}

func FIFOAntsRelease(wLtPool *WorkerFIFOPool) {
	err := wLtPool.ReleaseTimeout(time.Duration(releaseTimeout) * time.Second)
	if err != nil {
		wLtPool.Release()
	}
}

func LIFOAntsRelease(sLtPool *WorkerLIFOPool) {
	err := sLtPool.ReleaseTimeout(time.Duration(releaseTimeout) * time.Second)
	if err != nil {
		sLtPool.Release()
	}
}

func FIFOWorkerTaskSubmit(wLtPool *WorkerFIFOPool, fn func()) error {
	return wLtPool.Submit(fn)
}

func LIFOWorkerTaskSubmit(sLtPool *WorkerLIFOPool, fn func()) error {
	return sLtPool.Submit(fn)
}

func defaultWorkerPoolOpts() *ants.Options {
	return &ants.Options{
		ExpiryDuration:   expiryDuration,
		MaxBlockingTasks: maxBlockingTasks,
		PreAlloc:         preAlloc,
		Nonblocking:      nonblocking,
		PanicHandler: func(err interface{}) {
			logger.Errorf("goroutine pool panic: %v\n%s", err, debug.Stack())
		},
	}

}
