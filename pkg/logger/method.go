package logger

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	middleware "github.com/ishaqcherry9/depend/pkg/gin/middleware"
	"go.uber.org/zap"
)

func Debug(c *gin.Context, msg string, fields ...Field) {
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.GCtxRequestID(c)))
	getLogger().Debug(msg, fields...)
}

func Info(c *gin.Context, msg string, fields ...Field) {
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.GCtxRequestID(c)))
	getLogger().Info(msg, fields...)
}

func Warn(c *gin.Context, msg string, fields ...Field) {
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.GCtxRequestID(c)))
	getLogger().Warn(msg, fields...)
}

func Error(c *gin.Context, msg string, fields ...Field) {
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.GCtxRequestID(c)))
	getLogger().Error(msg, fields...)
}

func Panic(c *gin.Context, msg string, fields ...Field) {
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.GCtxRequestID(c)))
	getLogger().Panic(msg, fields...)
}

func Fatal(c *gin.Context, msg string, fields ...Field) {
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.GCtxRequestID(c)))
	getLogger().Fatal(msg, fields...)
}

func Debugf(c *gin.Context, format string, a ...interface{}) {
	msg, fields := getMessage(c, format, a...)
	getLogger().Debug(msg, fields...)
}

// getMessage format with Sprint, Sprintf, or neither.
func getMessage(c *gin.Context, template string, fmtArgs ...interface{}) (string, []Field) {
	fields := make([]Field, 0, 1)
	fields = append(fields, zap.String(middleware.ContextRequestIDKey, middleware.GCtxRequestID(c)))
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

func Infof(c *gin.Context, format string, a ...interface{}) {
	msg, fields := getMessage(c, format, a...)
	getLogger().Info(msg, fields...)
}

func Warnf(c *gin.Context, format string, a ...interface{}) {
	msg, fields := getMessage(c, format, a...)
	getSugaredLogger().Warnf(format, a...)
	getLogger().Warn(msg, fields...)
}

func Errorf(c *gin.Context, format string, a ...interface{}) {
	msg, fields := getMessage(c, format, a...)
	getLogger().Error(msg, fields...)
}

func Fatalf(c *gin.Context, format string, a ...interface{}) {
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
