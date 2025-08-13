package logger

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

var (
	defaultLevel    = "debug"
	defaultEncoding = formatConsole
	defaultIsSave   = false

	defaultFilename      = "out.log"
	defaultMaxSize       = 10
	defaultMaxBackups    = 100
	defaultMaxAge        = 30
	defaultIsCompression = false
	defaultIsLocalTime   = true
)

type options struct {
	level    string
	encoding string
	isSave   bool

	fileConfig *fileOptions

	hooks []func(zapcore.Entry) error
}

func defaultOptions() *options {
	return &options{
		level:    defaultLevel,
		encoding: defaultEncoding,
		isSave:   defaultIsSave,
	}
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

type Option func(*options)

func WithLevel(levelName string) Option {
	return func(o *options) {
		levelName = strings.ToUpper(levelName)
		switch levelName {
		case levelDebug, levelInfo, levelWarn, levelError:
			o.level = levelName
		default:
			o.level = levelDebug
		}
	}
}

func WithFormat(format string) Option {
	return func(o *options) {
		if strings.ToLower(format) == formatJSON {
			o.encoding = formatJSON
		}
	}
}

func WithSave(isSave bool, opts ...FileOption) Option {
	return func(o *options) {
		if isSave {
			o.isSave = true
			fo := defaultFileOptions()
			fo.apply(opts...)
			o.fileConfig = fo
		}
	}
}

func WithHooks(hooks ...func(zapcore.Entry) error) Option {
	return func(o *options) {
		o.hooks = hooks
	}
}

type fileOptions struct {
	filename      string
	maxSize       int
	maxBackups    int
	maxAge        int
	isCompression bool
	isLocalTime   bool
}

func defaultFileOptions() *fileOptions {
	return &fileOptions{
		filename:      defaultFilename,
		maxSize:       defaultMaxSize,
		maxBackups:    defaultMaxBackups,
		maxAge:        defaultMaxAge,
		isCompression: defaultIsCompression,
		isLocalTime:   defaultIsLocalTime,
	}
}

func (o *fileOptions) apply(opts ...FileOption) {
	for _, opt := range opts {
		opt(o)
	}
}

type FileOption func(*fileOptions)

func WithFileName(filename string) FileOption {
	return func(f *fileOptions) {
		if filename != "" {
			f.filename = filename
		}
	}
}

func WithFileMaxSize(maxSize int) FileOption {
	return func(f *fileOptions) {
		if maxSize > 0 {
			f.maxSize = maxSize
		}
	}
}

func WithFileMaxBackups(maxBackups int) FileOption {
	return func(f *fileOptions) {
		if f.maxBackups > 0 {
			f.maxBackups = maxBackups
		}
	}
}

func WithFileMaxAge(maxAge int) FileOption {
	return func(f *fileOptions) {
		if f.maxAge > 0 {
			f.maxAge = maxAge
		}
	}
}

func WithFileIsCompression(isCompression bool) FileOption {
	return func(f *fileOptions) {
		f.isCompression = isCompression
	}
}

func WithLocalTime(isLocalTime bool) FileOption {
	return func(f *fileOptions) {
		f.isLocalTime = isLocalTime
	}
}
