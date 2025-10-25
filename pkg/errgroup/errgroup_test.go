package errgroup

import (
	"context"
	"fmt"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGroup_Go(t *testing.T) {
	t.Run("successful goroutine", func(t *testing.T) {
		g, _ := WithContext(context.Background(), Options{MaxConcurrency: 1})
		var executed bool
		g.Go(func(ctx context.Context) error {
			executed = true
			return nil
		})
		err := g.Wait()
		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("failing goroutine", func(t *testing.T) {
		g, _ := WithContext(context.Background(), Options{MaxConcurrency: 1, Tag: "test-tag"})
		g.Go(func(ctx context.Context) error {
			return errors.New("test error")
		})
		err := g.Wait()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "goroutine with tag test-tag failed")
	})

	t.Run("multiple goroutines", func(t *testing.T) {
		g, _ := WithContext(context.Background(), Options{MaxConcurrency: 2})
		var counter int
		var mu sync.Mutex
		for i := 0; i < 5; i++ {
			g.Go(func(ctx context.Context) error {
				mu.Lock()
				counter++
				mu.Unlock()
				return nil
			})
		}
		err := g.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 5, counter)
	})

	t.Run("cancellation on error", func(t *testing.T) {
		g, _ := WithContext(context.Background(), Options{MaxConcurrency: 2})
		var executed int
		var mu sync.Mutex

		g.Go(func(ctx context.Context) error {
			mu.Lock()
			executed++
			mu.Unlock()
			return errors.New("first error")
		})

		g.Go(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(100 * time.Millisecond):
				mu.Lock()
				executed++
				mu.Unlock()
				return errors.New("second error") //Should never execute
			}
		})

		err := g.Wait()
		assert.Error(t, err)
		assert.Equal(t, 1, executed) // Only the first one should have run.
	})

	t.Run("max concurrency", func(t *testing.T) {
		maxConcurrency := 3
		g, _ := WithContext(context.Background(), Options{MaxConcurrency: maxConcurrency})
		start := time.Now()
		var running int
		var maxRunning int
		var mu sync.Mutex

		for i := 0; i < 5; i++ {
			g.Go(func(ctx context.Context) error {
				mu.Lock()
				running++
				if running > maxRunning {
					maxRunning = running
				}
				mu.Unlock()
				time.Sleep(50 * time.Millisecond) // Simulate work
				mu.Lock()
				running--
				mu.Unlock()
				return nil
			})
		}
		err := g.Wait()
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, duration, 200*time.Millisecond)
		assert.Equal(t, maxConcurrency, maxRunning)
	})

	t.Run("SysProcAttr", func(t *testing.T) {
		attr := &syscall.SysProcAttr{}
		g, _ := WithContext(context.Background(), Options{
			MaxConcurrency: 1,
			SysProcAttr:    attr,
		})

		assert.Equal(t, attr, g.attr)

		g.Go(func(ctx context.Context) error {
			return nil
		})
		assert.NoError(t, g.Wait())
	})

	t.Run("goroutine tag", func(t *testing.T) {
		tag := "test-goroutine"
		g, _ := WithContext(context.Background(), Options{
			MaxConcurrency: 1,
			Tag:            tag,
		})

		var capturedTag string
		g.Go(func(ctx context.Context) error {
			if t, ok := ctx.Value(goroutineTag).(string); ok {
				capturedTag = t
			}
			return nil
		})

		assert.NoError(t, g.Wait())
		assert.Equal(t, tag, capturedTag)
	})

	t.Run("multiple failing goroutines", func(t *testing.T) {
		g, _ := WithContext(context.Background(), Options{MaxConcurrency: 3, Tag: "multi-fail"})

		errorMessages := []string{"error 1", "error 2", "error 3"}
		for _, msg := range errorMessages {
			g.Go(func(ctx context.Context) error {
				return errors.New(msg)
			})
		}

		err := g.Wait()
		assert.Error(t, err)
		multierr, ok := err.(*multierror.Error)
		assert.True(t, ok)

		errs := multierr.Errors
		assert.Len(t, errs, 3)

		for i, e := range errs {
			assert.Contains(t, e.Error(), fmt.Sprintf("goroutine with tag multi-fail failed: %s", errorMessages[i]))
		}
	})
}

func TestWithContext(t *testing.T) {
	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		g, _ := WithContext(ctx, Options{MaxConcurrency: 1})

		var executed bool
		g.Go(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				executed = true
				return nil
			}
		})

		cancel()
		err := g.Wait()
		assert.Error(t, err)
		assert.False(t, executed)
	})
}

func TestGroup_Wait(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		g, _ := WithContext(context.Background(), Options{MaxConcurrency: 1})
		err := g.Wait()
		assert.NoError(t, err)
	})

	t.Run("one error", func(t *testing.T) {
		g, _ := WithContext(context.Background(), Options{MaxConcurrency: 1})
		g.Go(func(ctx context.Context) error {
			return errors.New("test error")
		})
		err := g.Wait()
		assert.Error(t, err)
	})

	t.Run("multiple errors", func(t *testing.T) {
		g, _ := WithContext(context.Background(), Options{MaxConcurrency: 2})
		g.Go(func(ctx context.Context) error {
			return errors.New("error 1")
		})
		g.Go(func(ctx context.Context) error {
			return errors.New("error 2")
		})
		err := g.Wait()
		assert.Error(t, err)
		multierr, ok := err.(*multierror.Error)
		assert.True(t, ok)
		assert.Len(t, multierr.Errors, 2)
	})
}

func TestGoroutineTagContext(t *testing.T) {
	t.Run("check tag value in context", func(t *testing.T) {
		tagValue := "test-tag"
		g, _ := WithContext(context.Background(), Options{MaxConcurrency: 1, Tag: tagValue})

		var capturedTag string
		g.Go(func(ctx context.Context) error {
			if value := ctx.Value(goroutineTag); value != nil {
				if tag, ok := value.(string); ok {
					capturedTag = tag
				}
			}
			return nil
		})

		err := g.Wait()
		assert.NoError(t, err)
		assert.Equal(t, tagValue, capturedTag)
	})

	t.Run("check no tag value in context", func(t *testing.T) {
		g, _ := WithContext(context.Background(), Options{MaxConcurrency: 1})

		var capturedValue interface{}
		g.Go(func(ctx context.Context) error {
			capturedValue = ctx.Value(goroutineTag)
			return nil
		})

		err := g.Wait()
		assert.NoError(t, err)
		assert.Equal(t, "", capturedValue)
	})
}
