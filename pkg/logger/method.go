package logger

import (
	"context"
	"fmt"
	"strings"

	middleware "github.com/ishaqcherry9/depend/pkg/gin/middleware"
	"go.uber.org/zap"
)

func Debug(c context.Context, msg string, fields ...Field) {
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.CtxRequestID(c)))
	getLogger().Debug(msg, fields...)
}

func Info(c context.Context, msg string, fields ...Field) {
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.CtxRequestID(c)))
	getLogger().Info(msg, fields...)
}

func Warn(c context.Context, msg string, fields ...Field) {
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.CtxRequestID(c)))
	getLogger().Warn(msg, fields...)
}

func Error(c context.Context, msg string, fields ...Field) {
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.CtxRequestID(c)))
	getLogger().Error(msg, fields...)
}

func Panic(c context.Context, msg string, fields ...Field) {
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.CtxRequestID(c)))
	getLogger().Panic(msg, fields...)
}

func Fatal(c context.Context, msg string, fields ...Field) {
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.CtxRequestID(c)))
	getLogger().Fatal(msg, fields...)
}

func Debugf(c context.Context, format string, a ...interface{}) {
	msg, fields := getMessage(c, format, a...)
	getLogger().Debug(msg, fields...)
}

// getMessage format with Sprint, Sprintf, or neither.
func getMessage(c context.Context, template string, fmtArgs ...interface{}) (string, []Field) {
	fields := make([]Field, 0, 1)
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.CtxRequestID(c)))
	if len(fmtArgs) == 0 {
		return template, fields
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...), fields
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str, fields
		}
	}
	return fmt.Sprint(fmtArgs...), fields
}

func Infof(c context.Context, format string, a ...interface{}) {
	msg, fields := getMessage(c, format, a...)
	getLogger().Info(msg, fields...)
}

func Warnf(c context.Context, format string, a ...interface{}) {
	msg, fields := getMessage(c, format, a...)
	getSugaredLogger().Warnf(format, a...)
	getLogger().Warn(msg, fields...)
}

func Errorf(c context.Context, format string, a ...interface{}) {
	msg, fields := getMessage(c, format, a...)
	getLogger().Error(msg, fields...)
}

func Fatalf(c context.Context, format string, a ...interface{}) {
	msg, fields := getMessage(c, format, a...)
	getLogger().Fatal(msg, fields...)
}

func Sync() error {
	_ = getSugaredLogger().Sync()
	err := getLogger().Sync()
	if err != nil && !strings.Contains(err.Error(), "/dev/stdout") {
		return err
	}
	return nil
}

func WithFields(fields ...Field) *zap.Logger {
	return GetWithSkip(0).With(fields...)
}
