package goredis

import (
	"crypto/tls"
	"time"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/sdk/trace"
)

type Option func(*options)

type options struct {
	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
	tlsConfig    *tls.Config

	singleOptions *redis.Options

	sentinelOptions *redis.FailoverOptions

	clusterOptions *redis.ClusterOptions

	enableTrace    bool
	tracerProvider *trace.TracerProvider
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

func defaultOptions() *options {
	return &options{
		enableTrace: false,
	}
}

func WithTracing(tp *trace.TracerProvider) Option {
	return func(o *options) {
		o.tracerProvider = tp
	}
}

func WithDialTimeout(t time.Duration) Option {
	return func(o *options) {
		o.dialTimeout = t
	}
}

func WithReadTimeout(t time.Duration) Option {
	return func(o *options) {
		o.readTimeout = t
	}
}

func WithWriteTimeout(t time.Duration) Option {
	return func(o *options) {
		o.writeTimeout = t
	}
}

func WithTLSConfig(c *tls.Config) Option {
	return func(o *options) {
		o.tlsConfig = c
	}
}

func WithSingleOptions(opt *redis.Options) Option {
	return func(o *options) {
		o.singleOptions = opt
	}
}

func WithSentinelOptions(opt *redis.FailoverOptions) Option {
	return func(o *options) {
		o.sentinelOptions = opt
	}
}

func WithClusterOptions(opt *redis.ClusterOptions) Option {
	return func(o *options) {
		o.clusterOptions = opt
	}
}
