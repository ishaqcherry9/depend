package captcha

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/ishaqcherry9/depend/pkg/errcode"
	"github.com/ishaqcherry9/depend/pkg/gin/middleware"
	"github.com/ishaqcherry9/depend/pkg/logger"
	yundun "github.com/yidun/yidun-golang-sdk/yidun/service/captcha"
)

type YunDunCheckParam struct {
	CaptchaId string `validate:"required"`
	SecretId  string `validate:"required"`
	SecretKey string `validate:"required"`
	Validate  string `validate:"required"`
	User      string
}

// 先检测返回值中error，err为nil时才有意义
func YunDunCheck(c *gin.Context, param YunDunCheckParam) (bool, *errcode.Error) {
	validate := validator.New()
	if err := validate.Struct(param); err != nil {
		return false, errcode.YunDunParamError
	}

	request := yundun.NewCaptchaVerifyRequest()
	request.SetCaptchaId(param.CaptchaId).SetValidate(param.Validate).SetUser(param.User)

	captchaClient := yundun.NewCaptchaVerifyClientWithAccessKey(param.SecretId, param.SecretKey)
	resp, err := captchaClient.Verify(request)
	if err != nil {
		logger.Errorf("[YunDun Req Err:%s][RequestId:%s]", err.Error(), middleware.GCtxRequestID(c))
		return false, errcode.YunDunRequestErr
	}

	logger.Infof("[YunDun Resp Err:%d][Msg:%s][Result:%t][RequestId:%s]", *resp.Error, *resp.Msg, *resp.Result)

	if *resp.Error != 0 {
		return false, errcode.YunDunResponseErr
	}

	return *resp.Result, nil
}
