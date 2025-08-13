package circuitbreaker

import (
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ishaqcherry9/depend/pkg/shield/window"
)

type Option func(*options)

const (
	StateOpen int32 = iota

	StateClosed
)

var (
	_ CircuitBreaker = &Breaker{}
)

type options struct {
	success float64
	request int64
	bucket  int
	window  time.Duration
}

func WithSuccess(s float64) Option {
	return func(c *options) {
		c.success = s
	}
}

func WithRequest(r int64) Option {
	return func(c *options) {
		c.request = r
	}
}

func WithWindow(d time.Duration) Option {
	return func(c *options) {
		c.window = d
	}
}

func WithBucket(b int) Option {
	return func(c *options) {
		c.bucket = b
	}
}

type Breaker struct {
	stat     window.RollingCounter
	r        *rand.Rand
	randLock sync.Mutex

	k       float64
	request int64

	state int32
}

func NewBreaker(opts ...Option) CircuitBreaker {
	opt := options{
		success: 0.6,
		request: 100,
		bucket:  10,
		window:  3 * time.Second,
	}
	for _, o := range opts {
		o(&opt)
	}
	counterOpts := window.RollingCounterOpts{
		Size:           opt.bucket,
		BucketDuration: time.Duration(int64(opt.window) / int64(opt.bucket)),
	}
	stat := window.NewRollingCounter(counterOpts)
	return &Breaker{
		stat:    stat,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
		request: opt.request,
		k:       1 / opt.success,
		state:   StateClosed,
	}
}

func (b *Breaker) summary() (success int64, total int64) {
	b.stat.Reduce(func(iterator window.Iterator) float64 {
		for iterator.Next() {
			bucket := iterator.Bucket()
			total += bucket.Count
			for _, p := range bucket.Points {
				success += int64(p)
			}
		}
		return 0
	})
	return
}

func (b *Breaker) Allow() error {
	accepts, total := b.summary()
	requests := b.k * float64(accepts)
	if total < b.request || float64(total) < requests {
		atomic.CompareAndSwapInt32(&b.state, StateOpen, StateClosed)
		return nil
	}
	atomic.CompareAndSwapInt32(&b.state, StateClosed, StateOpen)
	dr := math.Max(0, (float64(total)-requests)/float64(total+1))
	drop := b.trueOnProba(dr)
	if drop {
		return ErrNotAllowed
	}
	return nil
}

func (b *Breaker) MarkSuccess() {
	b.stat.Add(1)
}

func (b *Breaker) MarkFailed() {

	b.stat.Add(0)
}

func (b *Breaker) trueOnProba(proba float64) (truth bool) {
	b.randLock.Lock()
	truth = b.r.Float64() < proba
	b.randLock.Unlock()
	return truth
}
