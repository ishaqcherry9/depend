package errgroup

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGroup_Go(t *testing.T) {
	t.Run("successful goroutines", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		group := NewGroup(ctx, "test", 3)

		var result []int
		var mu sync.Mutex

		for i := 0; i < 5; i++ {
			taskID := i
			group.Go(func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					time.Sleep(time.Duration(taskID*100) * time.Millisecond) // Simulate work
					mu.Lock()
					result = append(result, taskID)
					mu.Unlock()
					return nil
				}
			})
		}

		err := group.Wait()
		assert.NoError(t, err)

		assert.Len(t, result, 5)
		assert.ElementsMatch(t, []int{0, 1, 2, 3, 4}, result)
	})

	t.Run("failed goroutine", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		group := NewGroup(ctx, "test", 3)

		var result []int
		var mu sync.Mutex

		failTaskID := 2

		for i := 0; i < 5; i++ {
			taskID := i
			group.Go(func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					if taskID == failTaskID {
						return errors.New("intentional failure")
					}
					time.Sleep(time.Duration(taskID*100) * time.Millisecond) // Simulate work
					mu.Lock()
					result = append(result, taskID)
					mu.Unlock()
					return nil
				}
			})
		}

		err := group.Wait()
		assert.Error(t, err)

		// 验证错误信息包含预期的失败任务的 tag
		assert.Contains(t, err.Error(), fmt.Sprintf("goroutine with tag test failed"))

		// 验证成功的任务数量和内容
		assert.Len(t, result, 4)
		assert.ElementsMatch(t, []int{0, 1, 3, 4}, result)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		group := NewGroup(ctx, "test", 3)

		var result []int
		var mu sync.Mutex

		// Cancel context after short delay
		time.AfterFunc(500*time.Millisecond, cancel)

		for i := 0; i < 5; i++ {
			taskID := i
			group.Go(func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					mu.Lock()
					result = append(result, taskID)
					mu.Unlock()
					return ctx.Err()
				case <-time.After(time.Second): // Simulate long running task
					mu.Lock()
					result = append(result, taskID)
					mu.Unlock()
					return nil
				}
			})
		}

		err := group.Wait()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")

		// 验证任务是否被取消
		assert.GreaterOrEqual(t, len(result), 1)
	})

	t.Run("concurrency limit", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		limit := 3
		group := NewGroup(ctx, "test", limit)

		var running int
		var maxRunning int
		var mu sync.Mutex

		task := func(ctx context.Context) error {
			mu.Lock()
			running++
			if running > maxRunning {
				maxRunning = running
			}
			mu.Unlock()

			select {
			case <-ctx.Done():
				mu.Lock()
				running--
				mu.Unlock()
				return ctx.Err()
			case <-time.After(500 * time.Millisecond):
				mu.Lock()
				running--
				mu.Unlock()
				return nil
			}
		}

		for i := 0; i < 5; i++ {
			group.Go(task)
		}

		err := group.Wait()
		assert.NoError(t, err)
		assert.Equal(t, limit, maxRunning)
	})
}

func Test_goroutineTag(t *testing.T) {
	ctx := context.Background()
	taggedCtx := context.WithValue(ctx, goroutineTag, "testTag")

	value := taggedCtx.Value(goroutineTag)
	assert.Equal(t, "testTag", value)

	emptyCtx := context.Background()
	emptyValue := emptyCtx.Value(goroutineTag)
	assert.Nil(t, emptyValue)
}
