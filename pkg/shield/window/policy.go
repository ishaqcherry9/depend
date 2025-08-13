package window

import (
	"sync"
	"time"
)

type RollingPolicy struct {
	mu     sync.RWMutex
	size   int
	window *Window
	offset int

	bucketDuration time.Duration
	lastAppendTime time.Time
}

type RollingPolicyOpts struct {
	BucketDuration time.Duration
}

func NewRollingPolicy(window *Window, opts RollingPolicyOpts) *RollingPolicy {
	return &RollingPolicy{
		window: window,
		size:   window.Size(),
		offset: 0,

		bucketDuration: opts.BucketDuration,
		lastAppendTime: time.Now(),
	}
}

func (r *RollingPolicy) timespan() int {
	v := int(time.Since(r.lastAppendTime) / r.bucketDuration)
	if v > -1 {
		return v
	}
	return r.size
}

func (r *RollingPolicy) apply(f func(offset int, val float64), val float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	timespan := r.timespan()
	oriTimespan := timespan
	if timespan > 0 {
		start := (r.offset + 1) % r.size
		end := (r.offset + timespan) % r.size
		if timespan > r.size {
			timespan = r.size
		}

		r.window.ResetBuckets(start, timespan)
		r.offset = end
		r.lastAppendTime = r.lastAppendTime.Add(time.Duration(oriTimespan * int(r.bucketDuration)))
	}
	f(r.offset, val)
}

func (r *RollingPolicy) Append(val float64) {
	r.apply(r.window.Append, val)
}

func (r *RollingPolicy) Add(val float64) {
	r.apply(r.window.Add, val)
}

func (r *RollingPolicy) Reduce(f func(Iterator) float64) (val float64) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	timespan := r.timespan()
	if count := r.size - timespan; count > 0 {
		offset := r.offset + timespan + 1
		if offset >= r.size {
			offset = offset - r.size
		}
		val = f(r.window.Iterator(offset, count))
	}
	return val
}
