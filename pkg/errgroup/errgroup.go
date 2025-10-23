package errgroup

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors" // 使用更高级的错误处理库
)

// 定义一个自定义的键类型，用于存储标签
type tagKey string

const goroutineTag tagKey = "GoroutineTag"

// GoFunc 定义要并发执行的函数类型，接收 context 并返回 error
type GoFunc func(context.Context) error

// Group 是一个自定义的 errgroup，具有并发控制、多错误收集和清晰的上下文管理功能。
type Group struct {
	sem    chan struct{}        // 控制并发数的信号量
	errMu  sync.Mutex           // 保护 result 的互斥锁
	result *multierror.Error    // 收集所有错误的 multierror 对象
	ctx    context.Context      // 用于控制所有 Goroutine 的生命周期
	cancel context.CancelFunc   // 取消函数
	wg     sync.WaitGroup       // 等待所有 Goroutine 完成
	attr   *syscall.SysProcAttr // 子进程的内核参数
	tag    string               // 协程标签
}

// Options 用于配置 Group 的选项
type Options struct {
	MaxConcurrency int                  // 最大并发数
	SysProcAttr    *syscall.SysProcAttr // 子进程的内核参数
	Tag            string               // 协程标签
}

// WithContext 创建一个带有上下文和选项的 Group。
func WithContext(ctx context.Context, options Options) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	g := &Group{
		sem:    make(chan struct{}, options.MaxConcurrency),
		ctx:    ctx,
		cancel: cancel,
		result: new(multierror.Error),
		attr:   options.SysProcAttr,
		tag:    options.Tag,
	}
	return g, ctx
}

// Go 启动一个新的 Goroutine，并限制并发数。
func (g *Group) Go(f GoFunc) {
	g.wg.Add(1)
	g.sem <- struct{}{} // 获取信号量
	go func() {
		defer func() {
			<-g.sem     // 释放信号量
			g.wg.Done() // 标记 Goroutine 完成
		}()

		// 创建带有标签的 context
		taggedCtx := context.WithValue(g.ctx, goroutineTag, g.tag)

		// 使用传入的 context 执行任务
		if err := f(taggedCtx); err != nil {
			g.errMu.Lock()
			g.result = multierror.Append(g.result, errors.Wrap(err, fmt.Sprintf("goroutine with tag %s failed", g.tag))) // 收集错误,添加上下文
			g.errMu.Unlock()
			g.cancel() // 发生错误时取消所有 Goroutine
		}
	}()
}

// Wait 等待所有 Goroutine 完成，并返回所有错误。
func (g *Group) Wait() error {
	g.wg.Wait()                  // 等待所有 Goroutine 完成
	g.cancel()                   // 确保所有 Goroutine 都被取消
	return g.result.ErrorOrNil() // 返回所有错误
}
