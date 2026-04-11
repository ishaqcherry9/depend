package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSetRequestHeadersContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("设置 Customeruid 与 member_id", func(t *testing.T) {
		router := gin.New()
		router.Use(SetRequestHeadersContext())
		router.GET("/test", func(c *gin.Context) {
			customerUID, ok := c.Get(HeaderCustomerUIDKey)
			if !ok {
				t.Fatalf("未设置 Customeruid")
			}
			if customerUID != "123" {
				t.Fatalf("Customeruid 不符合预期，want=%q, got=%v", "123", customerUID)
			}

			memberID, ok := c.Get(HeaderMemberIDKey)
			if !ok {
				t.Fatalf("未设置 member_id")
			}
			if memberID != "123" {
				t.Fatalf("member_id 不符合预期，want=%q, got=%v", "123", memberID)
			}

			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(HeaderCustomerUIDKey, "123")
		req.Header.Set(HeaderMemberIDKey, "123")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("响应状态码不符合预期，want=%d, got=%d", http.StatusOK, w.Code)
		}
	})

	t.Run("仅设置 Customeruid", func(t *testing.T) {
		router := gin.New()
		router.Use(SetRequestHeadersContext())
		router.GET("/test", func(c *gin.Context) {
			customerUID, ok := c.Get(HeaderCustomerUIDKey)
			if !ok {
				t.Fatalf("未设置 Customeruid")
			}
			if customerUID != "abc" {
				t.Fatalf("Customeruid 不符合预期，want=%q, got=%v", "abc", customerUID)
			}

			if _, ok := c.Get(HeaderMemberIDKey); ok {
				t.Fatalf("不应设置 member_id")
			}

			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(HeaderCustomerUIDKey, "abc")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("响应状态码不符合预期，want=%d, got=%d", http.StatusOK, w.Code)
		}
	})

	t.Run("不设置任何字段", func(t *testing.T) {
		router := gin.New()
		router.Use(SetRequestHeadersContext())
		router.GET("/test", func(c *gin.Context) {
			if _, ok := c.Get(HeaderCustomerUIDKey); ok {
				t.Fatalf("不应设置 Customeruid")
			}

			if _, ok := c.Get(HeaderMemberIDKey); ok {
				t.Fatalf("不应设置 member_id")
			}

			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("响应状态码不符合预期，want=%d, got=%d", http.StatusOK, w.Code)
		}
	})

	t.Run("设置 member_type 与 member_identity", func(t *testing.T) {
		router := gin.New()
		router.Use(SetRequestHeadersContext())
		router.GET("/test", func(c *gin.Context) {
			memberType, ok := c.Get(HeaderMemberTypeKey)
			if !ok {
				t.Fatalf("未设置 member_type")
			}
			if memberType != "2" {
				t.Fatalf("member_type 不符合预期，want=%q, got=%v", "2", memberType)
			}

			memberIdentity, ok := c.Get(HeaderMemberIdentityKey)
			if !ok {
				t.Fatalf("未设置 member_identity")
			}
			if memberIdentity != "9" {
				t.Fatalf("member_identity 不符合预期，want=%q, got=%v", "9", memberIdentity)
			}

			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(HeaderMemberTypeKey, "2")
		req.Header.Set(HeaderMemberIdentityKey, "9")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("响应状态码不符合预期，want=%d, got=%d", http.StatusOK, w.Code)
		}
	})
}
