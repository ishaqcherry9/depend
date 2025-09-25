package ctoken

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/ishaqcherry9/depend/pkg/goredis"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"strings"
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

	dataKV := gconv.Map(data)
	if dataKV == nil {
		err = gerror.NewCode(gcode.CodeMissingParameter, MsgErrPlatformEmpty)
		return
	}

	xDeviceId := gconv.String(dataKV["X-Device-Id"])
	if xDeviceId == "" {
		err = gerror.NewCode(gcode.CodeMissingParameter, MsgErrDeviceIDEmpty)
		return
	}
	xClient := gconv.String(dataKV["X-Client"])

	if xClient == "" {
		err = gerror.NewCode(gcode.CodeMissingParameter, MsgErrClientEmpty)
		return
	}

	var cacheKey string
	if c.Options.MultiLogin {

		cacheKey = fmt.Sprintf("%s:%s", xClient, userKey)
	} else {

		cacheKey = fmt.Sprintf("%s:%s:%s", xClient, xDeviceId, userKey)
	}

	if c.Options.MultiLogin {
		existingToken, _, getErr := c.Get(ctx, cacheKey)
		if getErr == nil && existingToken != "" {
			return existingToken, nil
		}
	}

	token, err = c.Codec.Encode(ctx, cacheKey)
	if err != nil {
		err = gerror.WrapCode(gcode.CodeInternalError, err)
		return
	}

	userCache := g.Map{
		KeyUserKey:    userKey,
		KeyXDeviceID:  xDeviceId,
		KeyXClient:    xClient,
		KeyToken:      token,
		KeyData:       data,
		KeyRefreshNum: 0,
		KeyCreateTime: gtime.Now().TimestampMilli(),
	}

	err = c.Cache.Set(ctx, cacheKey, userCache)
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

	cacheKey, err := c.Codec.Decrypt(ctx, token)
	if err != nil {
		err = gerror.WrapCode(gcode.CodeInvalidParameter, err)
		return
	}

	parts := strings.Split(cacheKey, ":")

	if c.Options.MultiLogin {

		if len(parts) != 2 {
			err = gerror.NewCode(gcode.CodeInvalidParameter, MsgErrTokenFormat)
			return
		}
		userKey = parts[1]
	} else {

		if len(parts) != 3 {
			err = gerror.NewCode(gcode.CodeInvalidParameter, MsgErrTokenFormat)
			return
		}
		userKey = parts[2]
	}

	userCache, err := c.Cache.Get(ctx, cacheKey)
	if err != nil {
		logger.Error("Validate cache.Get error", logger.Err(err), logger.String("cacheKey", cacheKey))
		return
	}
	if userCache == nil {
		logger.Warn("Validate userCache is nil", logger.String("userKey", userKey))
		err = gerror.NewCode(gcode.CodeInternalError, MsgErrDataEmpty)
		return
	}
	if token != gconv.String(userCache[KeyToken]) {
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
			err = c.Cache.Set(ctx, cacheKey, userCache)
			if err != nil {
				err = gerror.WrapCode(gcode.CodeInternalError, err)
				return
			}
		}
	}
	refreshToken()

	return
}

func (c *CToken) Get(ctx context.Context, cacheKey string) (token string, data any, err error) {
	if cacheKey == "" {
		err = gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
		return
	}

	userCache, err := c.Cache.Get(ctx, cacheKey)
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

	cacheKey, err := c.Codec.Decrypt(ctx, token)
	if err != nil {
		err = gerror.WrapCode(gcode.CodeInvalidParameter, err)
		return
	}

	// 解析出userKey
	parts := strings.Split(cacheKey, ":")
	if c.Options.MultiLogin {
		if len(parts) != 2 {
			return "", nil, gerror.NewCode(gcode.CodeInvalidParameter, MsgErrTokenFormat)
		}
		userKey = parts[1]
	} else {
		if len(parts) != 3 {
			return "", nil, gerror.NewCode(gcode.CodeInvalidParameter, MsgErrTokenFormat)
		}
		userKey = parts[2]
	}

	userCache, err := c.Cache.Get(ctx, cacheKey)
	if err != nil {
		return "", nil, gerror.WrapCode(gcode.CodeInternalError, err)
	}
	if userCache == nil {
		return "", nil, gerror.NewCode(gcode.CodeInternalError, MsgErrDataEmpty)
	}
	return userKey, userCache[KeyData], nil
}

func (c *CToken) Destroy(ctx context.Context, cacheKey string) error {
	if cacheKey == "" {
		return gerror.NewCode(gcode.CodeMissingParameter, MsgErrUserKeyEmpty)
	}

	parts := strings.Split(cacheKey, ":")
	if len(parts) != 3 {
		return gerror.NewCode(gcode.CodeInvalidParameter, "invalid key format")
	}
	client := parts[0]
	deviceId := parts[1]
	userKey := parts[2]

	var cacheMKey string
	if c.Options.MultiLogin {
		cacheMKey = fmt.Sprintf("%s:%s", client, userKey)
	} else {
		cacheMKey = fmt.Sprintf("%s:%s:%s", client, deviceId, userKey)
	}

	err := c.Cache.Remove(ctx, cacheMKey)
	if err != nil {
		return gerror.WrapCode(gcode.CodeInternalError, err)
	}
	return nil
}

func (c *CToken) GetOptions() Options {
	return c.Options
}
