package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return cors.New(
		cors.Config{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders: []string{
				"Origin", "Authorization", "Content-Type", "Accept",
				"X-Device-Id",  // 设备编码(必填)
				"X-Client",     // 终端标识(必填), 取值范围：h5 android ios pc
				"X-Token",      // 登录后颁发的token(必填)
				"X-Version",    // 客户端版本号(必填)
				"X-Timestamp",  // 发起请求的毫秒时间戳(必填)
				"X-RetryTimes", // 发起请求的第几次重试，如0、1、2、3
				"X-Timeout",    // 连接+读写超时总和，秒单位
			},
			ExposeHeaders:    []string{"Content-Length", "text/plain", "Authorization", "Content-Type", "X-Request-Id"},
			AllowCredentials: false,
			MaxAge:           12 * time.Hour,
		},
	)
}
