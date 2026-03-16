package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redis "github.com/redis/go-redis/v9"
)

func newPhpSessionTestConfig(t *testing.T) (PhpSessionConfig, func()) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("启动 miniredis 失败: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	cleanup := func() {
		_ = rdb.Close()
		mr.Close()
	}

	cfg := PhpSessionConfig{
		RedisClient: rdb,
		TokenConfig: PhpSessionTokenConfig{
			KeyVal: map[string]string{
				"pc": "pc",
			},
		},
		SiteID:                  "site_test",
		MemberLoginTokenHashKey: "member_login_token",
	}

	return cfg, cleanup
}

func TestPhpSession_WhitelistPathSkipAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg, cleanup := newPhpSessionTestConfig(t)
	defer cleanup()

	router := gin.New()
	router.Use(PhpSession(cfg, []string{"/public/ping"}))
	router.GET("/public/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/public/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("白名单路径应直接放行，want=%d, got=%d", http.StatusNoContent, w.Code)
	}
}

func TestPhpSession_NonWhitelistMissingHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg, cleanup := newPhpSessionTestConfig(t)
	defer cleanup()

	router := gin.New()
	router.Use(PhpSession(cfg, []string{"/public/ping"}))
	router.GET("/private/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/private/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("应返回业务状态码载体 HTTP 200，want=%d, got=%d", http.StatusOK, w.Code)
	}

	if body := w.Body.String(); body == "" {
		t.Fatalf("缺少请求头时应返回错误响应体")
	}
}

func TestIsPathWhitelisted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		whitelistPaths []string
		want           bool
	}{
		{
			name:           "精确匹配",
			path:           "/api/v1/ping",
			whitelistPaths: []string{"/api/v1/ping"},
			want:           true,
		},
		{
			name:           "前缀通配匹配",
			path:           "/api/v1/public/demo",
			whitelistPaths: []string{"/api/v1/public/*"},
			want:           true,
		},
		{
			name:           "不匹配",
			path:           "/api/v1/private/demo",
			whitelistPaths: []string{"/api/v1/public/*"},
			want:           false,
		},
		{
			name:           "忽略空白名单项",
			path:           "/api/v1/ping",
			whitelistPaths: []string{"", "/api/v1/health"},
			want:           false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isPathWhitelisted(tt.path, tt.whitelistPaths)
			if got != tt.want {
				t.Fatalf("匹配结果错误，want=%v, got=%v", tt.want, got)
			}
		})
	}
}
