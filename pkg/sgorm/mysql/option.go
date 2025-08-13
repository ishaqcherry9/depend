package mysql

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Option func(*options)

type options struct {
	isLog         bool
	slowThreshold time.Duration

	maxIdleConns    int
	maxOpenConns    int
	connMaxLifetime time.Duration

	disableForeignKey bool
	enableTrace       bool

	requestIDKey string
	gLog         *zap.Logger
	logLevel     logger.LogLevel

	slavesDsn  []string
	mastersDsn []string

	plugins []gorm.Plugin
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func defaultOptions() *options {
	return &options{
		isLog:           false,
		slowThreshold:   time.Duration(0),
		maxIdleConns:    3,
		maxOpenConns:    50,
		connMaxLifetime: 30 * time.Minute,

		disableForeignKey: true,
		enableTrace:       false,

		requestIDKey: "",
		gLog:         nil,
		logLevel:     logger.Info,
	}
}

func WithLogging(l *zap.Logger, level ...logger.LogLevel) Option {
	return func(o *options) {
		o.isLog = true
		o.gLog = l
		if len(level) > 0 {
			o.logLevel = level[0]
		} else {
			o.logLevel = logger.Info
		}
	}
}

func WithSlowThreshold(d time.Duration) Option {
	return func(o *options) {
		o.slowThreshold = d
	}
}

func WithMaxIdleConns(size int) Option {
	return func(o *options) {
		o.maxIdleConns = size
	}
}

func WithMaxOpenConns(size int) Option {
	return func(o *options) {
		o.maxOpenConns = size
	}
}

func WithConnMaxLifetime(t time.Duration) Option {
	return func(o *options) {
		o.connMaxLifetime = t
	}
}

func WithEnableForeignKey() Option {
	return func(o *options) {
		o.disableForeignKey = false
	}
}

func WithEnableTrace() Option {
	return func(o *options) {
		o.enableTrace = true
	}
}

func WithLogRequestIDKey(key string) Option {
	return func(o *options) {
		if key == "" {
			key = "request_id"
		}
		o.requestIDKey = key
	}
}

func WithRWSeparation(slavesDsn []string, mastersDsn ...string) Option {
	return func(o *options) {
		o.slavesDsn = slavesDsn
		o.mastersDsn = mastersDsn
	}
}

func WithGormPlugin(plugins ...gorm.Plugin) Option {
	return func(o *options) {
		o.plugins = plugins
	}
}
