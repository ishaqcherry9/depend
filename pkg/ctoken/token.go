package ctoken

import (
	"context"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/ishaqcherry9/depend/pkg/goredis"
	"github.com/ishaqcherry9/depend/pkg/logger"
)

type Token interface {
	Generate(ctx context.Context, userKey string, data any) (token string, err error)
	Validate(ctx context.Context, token string) (userKey string, err error)
	Get(ctx context.Context, userKey string) (token string, data any, err error)
	ParseToken(ctx context.Context, token string) (userKey string, data any, err error)
	Destroy(ctx context.Context, userKey string) error
	GetOptions() Options
}

type CToken struct {
	Options Options
	Codec   Codec
	Cache   Cache
}

func NewDefaultTokenByConfig(options Options, redisCli *goredis.Client) Token {
	if options.CacheMode == 0 {
		options.CacheMode = CacheModeCache
	}
	if options.CachePreKey == "" {
		options.CachePreKey = DefaultCacheKey
	}
	if options.Timeout == 0 {
		options.Timeout = DefaultTimeout
		options.MaxRefresh = DefaultTimeout / 2
	}
	if len(options.EncryptKey) == 0 {
		options.EncryptKey = []byte(DefaultEncryptKey)
	}
	if options.TokenDelimiter == "" {
		options.TokenDelimiter = DefaultTokenDelimiter
	}

	cToken := &CToken{
		Options: options,
		Codec:   NewDefaultCodec(options.TokenDelimiter, options.EncryptKey),
		Cache:   NewDefaultCache(options.CacheMode, options.CachePreKey, options.Timeout, redisCli),
	}
	logger.Debug("token options", logger.String("conf", options.String()))
	return cToken
}

func (c *CToken) Generate(ctx context.Context, userKey string, data any) (token string, err error) {
	if userKey == "" {
		err = gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
		return
	}

	if c.Options.MultiLogin {
		// 支持多端重复登录，如果获取到返回相同token
		token, _, err = c.Get(ctx, userKey)
		if err == nil && token != "" {
			return
		}
	}

	token, err = c.Codec.Encode(ctx, userKey)
	if err != nil {
		err = gerror.WrapCode(gcode.CodeInternalError, err)
		return
	}

	userCache := g.Map{
		KeyUserKey:    userKey,
		KeyToken:      token,
		KeyData:       data,
		KeyRefreshNum: 0,
		KeyCreateTime: gtime.Now().TimestampMilli(),
	}

	err = c.Cache.Set(ctx, userKey, userCache)
	if err != nil {
		err = gerror.WrapCode(gcode.CodeInternalError, err)
		return
	}

	return
}

func (c *CToken) Validate(ctx context.Context, token string) (userKey string, err error) {
	if token == "" {
		err = gerror.NewCode(gcode.CodeMissingParameter, MsgErrTokenEmpty)
		return
	}

	userKey, err = c.Codec.Decrypt(ctx, token)
	if err != nil {
		err = gerror.WrapCode(gcode.CodeInvalidParameter, err)
		return
	}
	userCache, err := c.Cache.Get(ctx, userKey)
	if err != nil {
		logger.Error("Validate cache.Get error", logger.Err(err), logger.String("userKey", userKey))
		return
	}
	if userCache == nil {
		logger.Warn("Validate userCache is nil", logger.String("userKey", userKey))
		err = gerror.NewCode(gcode.CodeInternalError, MsgErrDataEmpty)
		return
	}
	if token != userCache[KeyToken] {
		err = gerror.NewCode(gcode.CodeInvalidParameter, MsgErrValidate)
		return
	}

	logger.Debug("Validate userCache",
		logger.String("userKey", userKey),
		logger.Any("cache", userCache),
		logger.String("token", token))

	// 需要进行缓存超时时间刷新
	refreshToken := func() {
		nowTime := gtime.Now().TimestampMilli()
		createTime := userCache[KeyCreateTime]
		refreshNum := gconv.Int(userCache[KeyRefreshNum])
		if c.Options.MaxRefresh == 0 {
			return
		}
		if c.Options.MaxRefreshTimes > 0 && refreshNum >= c.Options.MaxRefreshTimes {
			return
		}
		if nowTime > gconv.Int64(createTime)+c.Options.MaxRefresh {
			userCache[KeyRefreshNum] = refreshNum + 1
			userCache[KeyCreateTime] = gtime.Now().TimestampMilli()
			err = c.Cache.Set(ctx, userKey, userCache)
			if err != nil {
				err = gerror.WrapCode(gcode.CodeInternalError, err)
				return
			}
		}
	}
	refreshToken()

	return
}

func (c *CToken) Get(ctx context.Context, userKey string) (token string, data any, err error) {
	if userKey == "" {
		err = gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
		return
	}

	userCache, err := c.Cache.Get(ctx, userKey)
	if err != nil {
		return "", nil, gerror.WrapCode(gcode.CodeInternalError, err)
	}
	if userCache == nil {
		return "", nil, gerror.NewCode(gcode.CodeInternalError, MsgErrDataEmpty)
	}
	return gconv.String(userCache[KeyToken]), userCache[KeyData], nil

}

func (c *CToken) ParseToken(ctx context.Context, token string) (userKey string, data any, err error) {
	if token == "" {
		err = gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
		return
	}

	userKey, err = c.Codec.Decrypt(ctx, token)
	if err != nil {
		err = gerror.WrapCode(gcode.CodeInvalidParameter, err)
		return
	}

	userCache, err := c.Cache.Get(ctx, userKey)
	if err != nil {
		return "", nil, gerror.WrapCode(gcode.CodeInternalError, err)
	}
	if userCache == nil {
		return "", nil, gerror.NewCode(gcode.CodeInternalError, MsgErrDataEmpty)
	}
	return userKey, userCache[KeyData], nil
}

func (c *CToken) Destroy(ctx context.Context, userKey string) error {
	if userKey == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}

	err := c.Cache.Remove(ctx, userKey)
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err)
	}
	return nil
}

func (c *CToken) GetOptions() Options {
	return c.Options
}
