package captcha

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/ishaqcherry9/depend/pkg/errcode"
	"github.com/ishaqcherry9/depend/pkg/logger"
	yundun "github.com/yidun/yidun-golang-sdk/yidun/service/captcha"
)

// 作为包级变量（线程安全），避免重复分配
var validate = validator.New()

type YunDunCheckParam struct {
	CaptchaId string `validate:"required"`
	SecretId  string `validate:"required"`
	SecretKey string `validate:"required"`
	Validate  string `validate:"required"`
	User      string
}

type YunDunResponse struct {
	Error  int    `json:"error"`
	Msg    string `json:"msg"`
	Result bool   `json:"result"`
}

// 先检测返回值中error，err为nil时才有意义
func YunDunCheck(ctx context.Context, param YunDunCheckParam) (YunDunResponse, *errcode.Error) {
	response := YunDunResponse{}

	if err := validate.Struct(param); err != nil {
		return response, errcode.YunDunParamError
	}

	request := yundun.NewCaptchaVerifyRequest()
	request.SetCaptchaId(param.CaptchaId).
		SetValidate(param.Validate).
		SetUser(param.User)

	captchaClient := yundun.NewCaptchaVerifyClientWithAccessKey(param.SecretId, param.SecretKey)
	resp, err := captchaClient.Verify(request)
	if err != nil {
		logger.Errorf(ctx, "[YunDun Req Err:%s]", err.Error())
		return response, errcode.YunDunRequestErr
	}
	if resp == nil {
		logger.Error(ctx, "[YunDun Resp nil]")
		return response, errcode.YunDunResponseErr
	}

	logger.Infof(ctx, "[YunDun Resp Err:%d][Msg:%s][Result:%t]", *resp.Error, *resp.Msg, *resp.Result)

	response.Error = *resp.Error
	response.Msg = *resp.Msg
	response.Result = *resp.Result

	return response, nil
}
