package logger

import (
	"strings"

	"go.uber.org/zap"
)

func Debug(msg string, fields ...Field) {
	getLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	getLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	getLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	getLogger().Error(msg, fields...)
}

func Panic(msg string, fields ...Field) {
	getLogger().Panic(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	getLogger().Fatal(msg, fields...)
}

func Debugf(format string, a ...interface{}) {
	getSugaredLogger().Debugf(format, a...)
}

func Infof(format string, a ...interface{}) {
	getSugaredLogger().Infof(format, a...)
}

func Warnf(format string, a ...interface{}) {
	getSugaredLogger().Warnf(format, a...)
}

func Errorf(format string, a ...interface{}) {
	getSugaredLogger().Errorf(format, a...)
}

func Fatalf(format string, a ...interface{}) {
	getSugaredLogger().Fatalf(format, a...)
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
