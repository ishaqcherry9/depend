package captchabase64

import (
	"context"
	"errors"
	"time"

	"image/color"

	"github.com/ishaqcherry9/depend/pkg/goredis"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"github.com/mojocn/base64Captcha"
	"go.uber.org/zap"
)

// 默认服务实例（可选）。
var captchaService *base64CaptchaService

// 初始化时确认入参,减少后期传参
type Base64CaptchaServiceConfig struct {
	Client      *goredis.Client // redis 客户端(必传)
	KeyPrefix   string          // redis 键前缀(可选,默认captcha:)
	Expiration  time.Duration   // redis 键过期时间(可选,默认10分钟)
	Width       int             // 图片宽度(可选,默认240)
	Height      int             // 图片高度(可选,默认80)
	NoiseCount  int             // 噪点数量(可选,默认50)
	DotCount    int             // 干扰点数量(可选,默认50)
	Length      int             // 数字验证码位数(可选,默认6)
	MaxSkew     float64         // 数字验证码扭曲程度(可选,默认0.6)
	Charset     string          // 字符串验证码字符集(可选,默认去除易混字符)
	LineOptions int             // 字符串验证码线条选项(可选,默认4)
	BgColor     *color.RGBA     // 背景色(可选)
}

// Service 组织验证码能力，持有存储实现。
type base64CaptchaService struct {
	store base64Captcha.Store
	*Base64CaptchaServiceConfig
}

// NewCaptchaService 使用默认参数构建客户端
func NewCaptchaService(ctx context.Context, client *goredis.Client) (*base64CaptchaService, error) {
	if client == nil {
		logger.Error(ctx, "captcha redis client is nil")
		return nil, errors.New("captcha redis client is nil")
	}
	return NewCaptchaServiceWithConfig(ctx, &Base64CaptchaServiceConfig{
		Client: client,
	})
}

// NewCaptchaServiceWithConfig 使用传入参数构建客户端
func NewCaptchaServiceWithConfig(ctx context.Context, config *Base64CaptchaServiceConfig) (*base64CaptchaService, error) {
	if config == nil || config.Client == nil {
		logger.Error(ctx, "captcha redis client is nil")
		return nil, errors.New("captcha redis client is nil")
	}
	// 设置默认值
	if config.KeyPrefix == "" {
		config.KeyPrefix = "captcha:"
	}
	if config.Expiration == 0 {
		config.Expiration = 10 * time.Minute
	}
	if config.Width == 0 {
		config.Width = 240
	}
	if config.Height == 0 {
		config.Height = 80
	}
	if config.NoiseCount == 0 {
		config.NoiseCount = 50
	}
	if config.DotCount == 0 {
		config.DotCount = 50
	}
	if config.Length == 0 {
		config.Length = 6
	}
	if config.MaxSkew == 0 {
		config.MaxSkew = 0.6
	}
	if config.Charset == "" {
		config.Charset = "123456789abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
	}
	if config.LineOptions == 0 {
		config.LineOptions = 4
	}
	captchaService = &base64CaptchaService{
		store:                      newRedisStore(config.Client, config.KeyPrefix, config.Expiration),
		Base64CaptchaServiceConfig: config,
	}
	return captchaService, nil
}

// 获取服务实例
func GetBase64CaptchaService() (*base64CaptchaService, error) {
	if captchaService == nil {
		logger.Error(context.Background(), "captcha service not initialized")
		return nil, errors.New("captcha service not initialized")
	}
	return captchaService, nil
}

// GenerateDigit 生成数字验证码，返回验证码 ID 与 base64 图片字符串。
// width/height 为图片尺寸，length 为验证码位数，maxSkew 扭曲程度(0.0~1.0)，dotCount 干扰点数量。
func (s *base64CaptchaService) GenerateDigit(ctx context.Context) (id string, b64Image string, err error) {
	if s == nil || s.store == nil {
		return "", "", errors.New("captcha service/store not initialized")
	}
	driver := base64Captcha.NewDriverDigit(s.Height, s.Width, s.Length, s.MaxSkew, s.DotCount)
	c := base64Captcha.NewCaptcha(driver, s.store)
	id, b64Image, answer, err := c.Generate()
	if err != nil {
		logger.Error(ctx, "captcha generate digit error", logger.String("error", err.Error()))
		return "", "", err
	}
	logger.Debug(ctx, "captcha generate digit", zap.String("id", id), zap.String("answer", answer))
	return
}

// GenerateDigitWith 生成数字验证码（可自定义参数，传0则回落到配置默认值）。
func (s *base64CaptchaService) GenerateDigitWith(ctx context.Context, width, height, length int, maxSkew float64, dotCount int) (id string, b64Image string, err error) {
	if s == nil || s.store == nil {
		return "", "", errors.New("captcha service/store not initialized")
	}
	if width <= 0 {
		width = s.Width
	}
	if height <= 0 {
		height = s.Height
	}
	if length <= 0 {
		length = s.Length
	}
	if maxSkew <= 0 {
		maxSkew = s.MaxSkew
	}
	if dotCount <= 0 {
		dotCount = s.DotCount
	}
	driver := base64Captcha.NewDriverDigit(height, width, length, maxSkew, dotCount)
	c := base64Captcha.NewCaptcha(driver, s.store)
	id, b64Image, answer, err := c.Generate()
	if err != nil {
		logger.Error(ctx, "captcha generate digit(with) error", logger.String("error", err.Error()))
		return "", "", err
	}
	logger.Debug(ctx, "captcha generate digit(with)", zap.String("id", id), zap.String("answer", answer))
	return
}

// GenerateMath 生成算术验证码，返回验证码 ID、base64 图片与答案。
// noiseCount 为噪点数量，bg/fonts 采用默认值。
func (s *base64CaptchaService) GenerateMath(ctx context.Context) (id string, b64Image string, err error) {
	if s == nil || s.store == nil {
		return "", "", errors.New("captcha service/store not initialized")
	}
	driver := base64Captcha.NewDriverMath(s.Height, s.Width, s.NoiseCount, 0, nil, nil, nil)
	c := base64Captcha.NewCaptcha(driver, s.store)
	id, b64Image, answer, err := c.Generate()
	if err != nil {
		logger.Error(ctx, "captcha generate math error", logger.String("error", err.Error()))
		return "", "", err
	}
	logger.Debug(ctx, "captcha GenerateMath", zap.String("id", id), zap.String("answer", answer))
	return
}

// GenerateString 生成字符串验证码，使用服务配置参数。
func (s *base64CaptchaService) GenerateString(ctx context.Context) (id string, b64Image string, err error) {
	if s == nil || s.store == nil {
		return "", "", errors.New("captcha service/store not initialized")
	}
	driver := base64Captcha.NewDriverString(
		s.Height,
		s.Width,
		s.Length,
		s.NoiseCount,
		s.LineOptions,
		s.Charset,
		s.BgColor,
		nil,
		nil,
	)
	c := base64Captcha.NewCaptcha(driver, s.store)
	id, b64Image, answer, err := c.Generate()
	if err != nil {
		logger.Error(ctx, "captcha generate string error", logger.String("error", err.Error()))
		return "", "", err
	}
	logger.Debug(ctx, "captcha GenerateString", zap.String("id", id), zap.String("answer", answer))
	return
}

// GenerateStringWith 生成字符串验证码（可自定义参数，传0或空值则回落到配置默认值）。
func (s *base64CaptchaService) GenerateStringWith(ctx context.Context, width, height, length, noiseCount, lineOptions int, charset string, bg *color.RGBA) (id string, b64Image string, err error) {
	if s == nil || s.store == nil {
		return "", "", errors.New("captcha service/store not initialized")
	}
	if width <= 0 {
		width = s.Width
	}
	if height <= 0 {
		height = s.Height
	}
	if length <= 0 {
		length = s.Length
	}
	if noiseCount <= 0 {
		noiseCount = s.NoiseCount
	}
	if lineOptions <= 0 {
		lineOptions = s.LineOptions
	}
	if charset == "" {
		charset = s.Charset
	}
	if bg == nil {
		bg = s.BgColor
	}
	driver := base64Captcha.NewDriverString(height, width, length, noiseCount, lineOptions, charset, bg, nil, nil)
	c := base64Captcha.NewCaptcha(driver, s.store)
	id, b64Image, answer, err := c.Generate()
	if err != nil {
		logger.Error(ctx, "captcha generate string(with) error", logger.String("error", err.Error()))
		return "", "", err
	}
	logger.Debug(ctx, "captcha GenerateString(with)", zap.String("id", id), zap.String("answer", answer))
	return
}

// GenerateMathWith 生成算术验证码（可自定义参数，传0则回落到配置默认值）。
func (s *base64CaptchaService) GenerateMathWith(ctx context.Context, width, height, noiseCount int) (id string, b64Image string, err error) {
	if s == nil || s.store == nil {
		return "", "", errors.New("captcha service/store not initialized")
	}
	if width <= 0 {
		width = s.Width
	}
	if height <= 0 {
		height = s.Height
	}
	if noiseCount <= 0 {
		noiseCount = s.NoiseCount
	}
	driver := base64Captcha.NewDriverMath(height, width, noiseCount, 0, nil, nil, nil)
	c := base64Captcha.NewCaptcha(driver, s.store)
	id, b64Image, answer, err := c.Generate()
	if err != nil {
		logger.Error(ctx, "captcha generate math(with) error", logger.String("error", err.Error()))
		return "", "", err
	}
	logger.Debug(ctx, "captcha GenerateMath(with)", zap.String("id", id), zap.String("answer", answer))
	return
}

// Verify 校验验证码，clearOnVerify 为 true 时校验成功后清除。
func (s *base64CaptchaService) Verify(ctx context.Context, id, value string, clearOnVerify bool) bool {
	if s == nil || s.store == nil {
		logger.Error(ctx, "captcha service/store not initialized")
		return false
	}
	return s.store.Verify(id, value, clearOnVerify)
}
