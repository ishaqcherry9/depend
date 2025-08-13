package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Field = zapcore.Field

func Int(key string, val int) Field {
	return zap.Int(key, val)
}

func Int32(key string, val int32) Field {
	return zap.Int32(key, val)
}

func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

func Uint(key string, val uint) Field {
	return zap.Uint(key, val)
}

func Uint32(key string, val uint32) Field {
	return zap.Uint32(key, val)
}

func Uint64(key string, val uint64) Field {
	return zap.Uint64(key, val)
}

func Uintptr(key string, val uintptr) Field {
	return zap.Uintptr(key, val)
}

func Float64(key string, val float64) Field {
	return zap.Float64(key, val)
}

func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

func String(key string, val string) Field {
	return zap.String(key, val)
}

func ByteString(key string, val []byte) Field {
	return zap.ByteString(key, val)
}

func Stringer(key string, val fmt.Stringer) Field {
	return zap.Stringer(key, val)
}

func Time(key string, val time.Time) Field {
	return zap.Time(key, val)
}

func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

func Err(err error) Field {
	return zap.Error(err)
}

func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}
