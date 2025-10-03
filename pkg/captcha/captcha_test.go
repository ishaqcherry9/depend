package captcha

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"testing"
)

// go test -v -test.run TestYunDunCheck
func TestYunDunCheck(t *testing.T) {
	validate := "VnQ-P4nETS4ToEP6w0ZQIhAbjyxPod1fceKXkxScvMSkkH4ider5prw8BVafKBP79z4MBNRacPo.qPjeE.VJu6-MpOOIueY2zoA4aEHf5EM9Eh7RN6UKRb-BGh09H8CD-9HFX7rgwOmviPquD4ioVnsYwMTaP8LioFXZx26nLCzl6adYCW0rIQdMA17p4rSF71or5R66i4rSLRNN5Tqy8BzO70XtxA2LpOeK0TcfPgavHrYwocvxCPRDQN4XDjdAfTJMwbLGqYQavO26j01gIeTEQ_8WAqCgkw6JdwofxYg6Ni.YixIqlmYDu92FcfPZP-0NFmKU8KBJIFHTLYnzxQPiFMOQ0FvhnZtJPR6IkM4Z6deBkcbw.eo5TaxGQ2x7DAjhYZh-riegT2iT8EjLxIUpwY6.h989LrnoB.QCc50QbN2EgLbuikh.tqVIZHQUOC7pIQ9XKIVjpp5HObZPSc-.QiU1-8VYIi-P8Jy-MQ0hvHQP-GjhwR8yXZa3"
	//测试key
	param := YunDunCheckParam{
		CaptchaId: "2558575287f34303abbffa7f0d92eabb",
		SecretId:  "09073883f8e4de28d4299171523193d0",
		SecretKey: "a6d778278626b54abe4c6df82d2ccdce",
		Validate:  validate,
	}

	c := gin.Context{}
	resultBool, err := YunDunCheck(&c, param)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	fmt.Println(resultBool)
}
