package logger

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// ContextRequestIDKey 请求ID的key
var ContextRequestIDKey = "request_id"

func Debug(c context.Context, msg string, fields ...Field) {
	fields = append(fields, AddCommonFields(c)...)
	getLogger().Debug(msg, fields...)
}

func Info(c context.Context, msg string, fields ...Field) {
	fields = append(fields, AddCommonFields(c)...)
	getLogger().Info(msg, fields...)
}

func Warn(c context.Context, msg string, fields ...Field) {
	fields = append(fields, AddCommonFields(c)...)
	getLogger().Warn(msg, fields...)
}

func Error(c context.Context, msg string, fields ...Field) {
	fields = append(fields, AddCommonFields(c)...)
	getLogger().Error(msg, fields...)
}

func Panic(c context.Context, msg string, fields ...Field) {
	fields = append(fields, AddCommonFields(c)...)
	getLogger().Panic(msg, fields...)
}

func Fatal(c context.Context, msg string, fields ...Field) {
	fields = append(fields, AddCommonFields(c)...)
	getLogger().Fatal(msg, fields...)
}

func Debugf(c context.Context, format string, a ...interface{}) {
	msg, fields := getMessage(c, format, a...)
	getLogger().Debug(msg, fields...)
}

// getMessage format with Sprint, Sprintf, or neither.
func getMessage(c context.Context, template string, fmtArgs ...interface{}) (string, []Field) {
	fields := make([]Field, 0)
	// fields = append(fields, AddCommonFields(c)...)
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
	Info(c, msg, fields...)
}

func Warnf(c context.Context, format string, a ...interface{}) {
	msg, fields := getMessage(c, format, a...)
	Warn(c, msg, fields...)
}

func Errorf(c context.Context, format string, a ...interface{}) {
	msg, fields := getMessage(c, format, a...)
	Error(c, msg, fields...)
}

func Fatalf(c context.Context, format string, a ...interface{}) {
	msg, fields := getMessage(c, format, a...)
	Fatal(c, msg, fields...)
}

func Sync() error {
	_ = getSugaredLogger().Sync()
	err := getLogger().Sync()
	if err != nil && !strings.Contains(err.Error(), "/dev/stdout") {
		return err
	}
	return nil
}

// 添加公共字段到日志中
func AddCommonFields(ctx context.Context) []Field {
	fields := make([]Field, 0, 1)
	// 请求ID
	fields = append(fields, zap.String(ContextRequestIDKey, CtxRequestID(ctx)))
	return fields
}

func WithFields(fields ...Field) *zap.Logger {
	return GetWithSkip(0).With(fields...)
}

func CtxRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	// // gin框架context
	// if c, ok := ctx.(*gin.Context); ok {
	// 	return GCtxRequestID(c)
	// }
	// 原生context,兼容gin框架context
	v := ctx.Value(ContextRequestIDKey)
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}
