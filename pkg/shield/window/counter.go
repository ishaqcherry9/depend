package window

import (
	"fmt"
	"time"
)

type Metric interface {
	Add(int64)

	Value() int64
}

type Aggregation interface {
	Min() float64

	Max() float64

	Avg() float64

	Sum() float64
}

type RollingCounter interface {
	Metric
	Aggregation

	Timespan() int

	Reduce(func(Iterator) float64) float64
}

type RollingCounterOpts struct {
	Size           int
	BucketDuration time.Duration
}

type rollingCounter struct {
	policy *RollingPolicy
}

func NewRollingCounter(opts RollingCounterOpts) RollingCounter {
	window := NewWindow(Options{Size: opts.Size})
	policy := NewRollingPolicy(window, RollingPolicyOpts{BucketDuration: opts.BucketDuration})
	return &rollingCounter{
		policy: policy,
	}
}

func (r *rollingCounter) Add(val int64) {
	if val < 0 {
		panic(fmt.Errorf("stat/metric: cannot decrease in value. val: %d", val))
	}
	r.policy.Add(float64(val))
}

func (r *rollingCounter) Reduce(f func(Iterator) float64) float64 {
	return r.policy.Reduce(f)
}

func (r *rollingCounter) Avg() float64 {
	return r.policy.Reduce(Avg)
}

func (r *rollingCounter) Min() float64 {
	return r.policy.Reduce(Min)
}

func (r *rollingCounter) Max() float64 {
	return r.policy.Reduce(Max)
}

func (r *rollingCounter) Sum() float64 {
	return r.policy.Reduce(Sum)
}

func (r *rollingCounter) Value() int64 {
	return int64(r.Sum())
}

func (r *rollingCounter) Timespan() int {
	r.policy.mu.RLock()
	defer r.policy.mu.RUnlock()
	return r.policy.timespan()
}
