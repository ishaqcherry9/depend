package goredis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/ishaqcherry9/depend/pkg/gin/middleware"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// NewZapLoggerHook 初始化redis hook。日志hook，集成zap log
func NewZapLoggerHook(logger *zap.Logger) redis.Hook {
	return &zapLoggerHook{
		Logger: logger,
	}
}

// zapLoggerHook redis 日志hook，集成zap log
type zapLoggerHook struct {
	Logger *zap.Logger
}

func (l *zapLoggerHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}
func (l *zapLoggerHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		// 开始时间，毫秒
		startTime := time.Now().UnixNano() / int64(time.Millisecond)
		err := next(ctx, cmd)
		endTime := time.Now().UnixNano() / int64(time.Millisecond)
		cost := endTime - startTime
		logger := l.Logger.With(zap.String(middleware.ContextRequestIDKey, middleware.CtxRequestID(ctx)))
		if cmd.Err() == nil || errors.Is(cmd.Err(), redis.Nil) {
			logger.Info(cmd.String(), zap.String("cost", fmt.Sprintf("%dms", cost)))
		} else {
			logger.Error(cmd.String(), zap.String("cost", fmt.Sprintf("%dms", cost)), zap.Error(cmd.Err()))
		}
		return err
	}
}
func (l *zapLoggerHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}
