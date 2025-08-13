package goredis

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

type Client = redis.Client

const (
	ErrRedisNotFound = redis.Nil

	DefaultRedisName = "default"
)

func Init(dsn string, opts ...Option) (*redis.Client, error) {
	o := defaultOptions()
	o.apply(opts...)

	opt, err := getRedisOpt(dsn, o)
	if err != nil {
		return nil, err
	}

	if o.singleOptions != nil {
		opt = o.singleOptions
	}

	rdb := redis.NewClient(opt)

	if o.tracerProvider != nil {
		err = redisotel.InstrumentTracing(rdb, redisotel.WithTracerProvider(o.tracerProvider))
		if err != nil {
			return nil, err
		}
	}

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	err = rdb.Ping(ctx).Err()

	return rdb, err
}

func InitSingle(addr string, password string, db int, opts ...Option) (*redis.Client, error) {
	o := defaultOptions()
	o.apply(opts...)

	opt := &redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  o.dialTimeout,
		ReadTimeout:  o.readTimeout,
		WriteTimeout: o.writeTimeout,
		TLSConfig:    o.tlsConfig,
	}

	if o.singleOptions != nil {
		opt = o.singleOptions
	}

	rdb := redis.NewClient(opt)

	if o.tracerProvider != nil {
		err := redisotel.InstrumentTracing(rdb, redisotel.WithTracerProvider(o.tracerProvider))
		if err != nil {
			return nil, err
		}
	}

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	err := rdb.Ping(ctx).Err()

	return rdb, err
}

func InitSentinel(masterName string, addrs []string, username string, password string, opts ...Option) (*redis.Client, error) {
	o := defaultOptions()
	o.apply(opts...)

	opt := &redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: addrs,
		Username:      username,
		Password:      password,
		DialTimeout:   o.dialTimeout,
		ReadTimeout:   o.readTimeout,
		WriteTimeout:  o.writeTimeout,
		TLSConfig:     o.tlsConfig,
	}

	if o.sentinelOptions != nil {
		opt = o.sentinelOptions
	}

	rdb := redis.NewFailoverClient(opt)

	if o.tracerProvider != nil {
		err := redisotel.InstrumentTracing(rdb, redisotel.WithTracerProvider(o.tracerProvider))
		if err != nil {
			return nil, err
		}
	}

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	err := rdb.Ping(ctx).Err()

	return rdb, err
}

func InitCluster(addrs []string, username string, password string, opts ...Option) (*redis.ClusterClient, error) {
	o := defaultOptions()
	o.apply(opts...)

	opt := &redis.ClusterOptions{
		Addrs:        addrs,
		Username:     username,
		Password:     password,
		DialTimeout:  o.dialTimeout,
		ReadTimeout:  o.readTimeout,
		WriteTimeout: o.writeTimeout,
		TLSConfig:    o.tlsConfig,
	}

	if o.clusterOptions != nil {
		opt = o.clusterOptions
	}

	clusterRdb := redis.NewClusterClient(opt)

	if o.tracerProvider != nil {
		err := redisotel.InstrumentTracing(clusterRdb, redisotel.WithTracerProvider(o.tracerProvider))
		if err != nil {
			return nil, err
		}
	}

	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	err := clusterRdb.ForEachMaster(ctx, func(ctx context.Context, client *redis.Client) error {
		return client.Ping(ctx).Err()
	})

	return clusterRdb, err
}

func getRedisOpt(dsn string, opts *options) (*redis.Options, error) {
	dsn = strings.ReplaceAll(dsn, " ", "")
	if len(dsn) > 8 {
		if !strings.Contains(dsn[len(dsn)-3:], "/") {
			dsn += "/0"
		}

		if dsn[:8] != "redis://" && dsn[:9] != "rediss://" {
			dsn = "redis://" + dsn
		}
	}

	redisOpts, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, err
	}

	if opts.dialTimeout > 0 {
		redisOpts.DialTimeout = opts.dialTimeout
	}
	if opts.readTimeout > 0 {
		redisOpts.ReadTimeout = opts.readTimeout
	}
	if opts.writeTimeout > 0 {
		redisOpts.WriteTimeout = opts.writeTimeout
	}
	if opts.tlsConfig != nil {
		redisOpts.TLSConfig = opts.tlsConfig
	}

	return redisOpts, nil
}

func Close(rdb *redis.Client) error {
	if rdb == nil {
		return nil
	}

	err := rdb.Close()
	if err != nil && errors.Is(err, redis.ErrClosed) {
		return err
	}

	return nil
}

func CloseCluster(clusterRdb *redis.ClusterClient) error {
	if clusterRdb == nil {
		return nil
	}

	err := clusterRdb.Close()
	if err != nil && errors.Is(err, redis.ErrClosed) {
		return err
	}

	return nil
}
