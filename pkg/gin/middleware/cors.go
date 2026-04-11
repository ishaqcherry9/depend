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
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
			AllowHeaders: []string{
				"Origin", "Authorization", "Content-Type", "Accept",
				"X-Device-Id",  // 设备编码(必填)
				"X-Client",     // 终端标识(必填), 取值范围：h5 android ios pc
				"X-Token",      // 登录后颁发的token(必填)
				"X-Version",    // 客户端版本号(必填)
				"X-Timestamp",  // 发起请求的毫秒时间戳(必填)
				"X-RetryTimes", // 发起请求的第几次重试，如0、1、2、3
				"X-Timeout",    // 连接+读写超时总和，秒单位
				"customeruid", "device", "token", "timestamp", "Cache-Control", "Pragma", "Priority", "Sec-CH-UA", "Sec-CH-UA-Mobile", "Sec-CH-UA-Platform", // 前台与浏览器跨域请求头
				"member_id", "member_type", "member_identity", "X-Request-Id", // 其他拦截器使用的请求头
				"DNT", "If-Modified-Since", "Keep-Alive", "User-Agent", "X-Mx-ReqToken", "X-Requested-With", "signature", "identifier", "version", "versionCode", "device-id", "customerUID", "stageToken", // 对齐参考CORS请求头
			},
			ExposeHeaders: []string{
				"Content-Length", "text/plain", "Authorization", "Content-Type", "X-Request-Id",
				"X-Trace-Id", "Date", "Server", "X-Powered-By", // 新增响应头，单独一行追加
			},
			AllowCredentials: false,
			MaxAge:           12 * time.Hour,
		},
	)
}
