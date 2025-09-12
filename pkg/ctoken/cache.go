package ctoken

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/ishaqcherry9/depend/pkg/goredis"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"time"
)

type Cache interface {
	Set(ctx context.Context, cacheKey string, cacheValue g.Map) error
	Get(ctx context.Context, cacheKey string) (g.Map, error)
	Remove(ctx context.Context, cacheKey string) error
}

type DefaultCache struct {
	Cache   *gcache.Cache
	Mode    int8
	PreKey  string
	Timeout int64
}

func NewDefaultCache(mode int8, preKey string, timeout int64, redisCli *goredis.Client) *DefaultCache {
	c := &DefaultCache{
		Cache:   gcache.New(),
		Mode:    mode,
		PreKey:  preKey,
		Timeout: timeout,
	}

	if c.Mode == CacheModeFile {
		c.initFileCache(gctx.New())
	} else if c.Mode == CacheModeRedis {
		c.Cache.SetAdapter(NewRedisAdapter(redisCli))
	}

	return c
}

func (c *DefaultCache) Set(ctx context.Context, cacheKey string, cacheValue g.Map) error {
	if cacheValue == nil {
		return errors.New(MsgErrDataEmpty)
	}
	value, err := gjson.Encode(cacheValue)
	if err != nil {
		return err
	}
	err = c.Cache.Set(ctx, c.PreKey+cacheKey, string(value), gconv.Duration(c.Timeout)*time.Millisecond)
	if err != nil {
		return err
	}
	if c.Mode == CacheModeFile {
		c.writeFileCache(ctx)
	}
	return nil
}

func (c *DefaultCache) Get(ctx context.Context, cacheKey string) (g.Map, error) {
	dataVar, err := c.Cache.Get(ctx, c.PreKey+cacheKey)
	if err != nil {
		logger.Error("cache.Get error", logger.Err(err), logger.String("key", c.PreKey+cacheKey))
		return nil, err
	}
	if dataVar.IsNil() {
		logger.Warn("cache.Get dataVar is nil", logger.String("key", c.PreKey+cacheKey))
		return nil, nil
	}
	//return dataVar.Map(), nil
	// 打印看看实际获取到的是什么
	logger.Debug("cache.Get raw data",
		logger.String("key", c.PreKey+cacheKey),
		logger.String("type", dataVar.String()),
		logger.Any("value", dataVar.Val()))

	result := dataVar.Map()
	logger.Debug("cache.Get converted map",
		logger.String("key", c.PreKey+cacheKey),
		logger.Any("map", result))

	return result, nil
}

func (c *DefaultCache) Remove(ctx context.Context, cacheKey string) error {
	_, err := c.Cache.Remove(ctx, c.PreKey+cacheKey)
	if c.Mode == CacheModeFile {
		c.writeFileCache(ctx)
	}
	return err
}

func (c *DefaultCache) initFileCache(ctx context.Context) {
	fileName := gstr.Replace(c.PreKey, ":", "_") + CacheModeFileDat
	file := gfile.Temp(fileName)
	logger.Debug("file cache init", logger.String("file", file))
	if !gfile.Exists(file) {
		return
	}
	data := gfile.GetContents(file)
	maps := gconv.Map(data)
	if maps == nil || len(maps) <= 0 {
		return
	}
	for k, v := range maps {
		_ = c.Cache.Set(ctx, k, v, gconv.Duration(c.Timeout)*time.Millisecond)
	}
}

func (c *DefaultCache) writeFileCache(ctx context.Context) {
	fileName := gstr.Replace(c.PreKey, ":", "_") + CacheModeFileDat
	file := gfile.Temp(fileName)
	data, e := c.Cache.Data(ctx)
	if e != nil {
		logger.Error("[CToken]cache writeFileCache data error", logger.Err(e))
	}
	e = gfile.PutContents(file, gjson.New(data).MustToJsonString())
	if e != nil {
		logger.Error("[CToken]cache writeFileCache put error", logger.Err(e))
	}
}
