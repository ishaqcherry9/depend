package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

const (
	// 客户ID
	HeaderCustomerUIDKey    = "Customeruid"
	HeaderMemberIDKey       = "member_id"
	HeaderMemberTypeKey     = "member_type"
	HeaderMemberIdentityKey = "member_identity"
)

// SetRequestHeadersContext 将请求头中的通用字段填充到 gin 上下文中。
func SetRequestHeadersContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 将 customer_uid 设置到 gin 上下文中
		if customerUID := c.GetHeader(HeaderCustomerUIDKey); customerUID != "" {
			c.Set(HeaderCustomerUIDKey, customerUID)
		}
		// 将 member_id 设置到 gin 上下文中
		if memberID := c.GetHeader(HeaderMemberIDKey); memberID != "" {
			c.Set(HeaderMemberIDKey, memberID)
		}
		// 将 member_type 设置到 gin 上下文中
		if memberType := c.GetHeader(HeaderMemberTypeKey); memberType != "" {
			c.Set(HeaderMemberTypeKey, memberType)
		}
		// 将 member_identity 设置到 gin 上下文中
		if memberIdentity := c.GetHeader(HeaderMemberIdentityKey); memberIdentity != "" {
			c.Set(HeaderMemberIdentityKey, memberIdentity)
		}

		c.Next()
	}
}

// 从上下文中获取用户ID。
func GetMemberIdByCtx(c context.Context) string {
	memberID := c.Value(HeaderMemberIDKey)
	if memberIDStr, ok := memberID.(string); ok {
		return memberIDStr
	}
	return ""
}

// 从上下文中获取客户ID。
func GetCustomerUidByCtx(c context.Context) string {
	customerUID := c.Value(HeaderCustomerUIDKey)
	if customerUIDStr, ok := customerUID.(string); ok {
		return customerUIDStr
	}
	return ""
}
