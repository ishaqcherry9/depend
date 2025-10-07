# depend
基础lib库



# 图形验证码简单使用方法

```
import "github.com/ishaqcherry9/depend/pkg/captcha/captchabase64"

// 初始化图形验证码客户端
	_, err = captchabase64.NewCaptchaService(context.Background(), database.GetRedisCli())
	if err != nil {
		logger.Error(context.Background(), "init captchabase64Service failed", zap.Error(err))
		return
	}

// 获取图形验证码方法
    captchabase64Service, err := captchabase64.GetBase64CaptchaService()
    if err != nil {
        logger.Error(c, "GetBase64CaptchaService", logger.Err(err))
        return
    }
    id, b64Image, answer, err := captchabase64Service.GenerateMath(c)
    if err != nil {
        logger.Error(c, "GenerateMath", logger.Err(err))
        return
    }

// 验证图形验证码
    if !captchabase64Service.Verify(c, id, answer, true) {
        logger.Error(c, "Verify", logger.Err(errors.New("验证图形验证码失败")))
        return
    }
```

