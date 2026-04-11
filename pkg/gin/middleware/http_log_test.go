package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHTTPLogMiddleware_ResponseWriteNoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(HttpLogMiddleware())
	router.GET("/http-log", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "ok",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/http-log", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("响应状态码不符合预期，want=%d, got=%d", http.StatusOK, w.Code)
	}
	if body := w.Body.String(); body == "" {
		t.Fatalf("响应体不应为空")
	}
}
