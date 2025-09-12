package ctoken

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/ishaqcherry9/depend/pkg/errcode"
	"github.com/ishaqcherry9/depend/pkg/gin/response"
	"strings"
)

type CTMiddleware struct {
	Token  Token
	ResFun func(c *gin.Context, err error)
}

func NewCTMiddleware(token Token) CTMiddleware {
	return CTMiddleware{
		Token: token,
		ResFun: func(c *gin.Context, err error) {
			response.Error(c, errcode.Unauthorized.RewriteMsg(MsgErrAuthInvalid))
		},
	}
}

func (c CTMiddleware) Auth() gin.HandlerFunc {
	return func(g *gin.Context) {
		if c.HasExcludePath(g) {
			g.Next()
			return
		}

		token, err := GetRequestToken(g)
		if err != nil {
			c.ResFun(g, err)
			g.Abort()
			return
		}
		userKey, err := c.Token.Validate(g.Request.Context(), token)
		if err != nil {
			c.ResFun(g, err)
			g.Abort()
			return
		}
		g.Set(KeyUserKey, userKey)
		g.Next()
	}
}

func (c CTMiddleware) HasExcludePath(g *gin.Context) bool {
	var (
		urlPath      = g.Request.URL.Path
		excludePaths = c.Token.GetOptions().AuthExcludePaths
	)
	if len(excludePaths) == 0 {
		return false
	}

	if strings.HasSuffix(urlPath, "/") {
		urlPath = gstr.SubStr(urlPath, 0, len(urlPath)-1)
	}

	for _, excludePath := range excludePaths {
		tmpPath := excludePath
		if strings.HasSuffix(tmpPath, "/*") {
			tmpPath = gstr.SubStr(tmpPath, 0, len(tmpPath)-2)
			if gstr.HasPrefix(urlPath, tmpPath) {
				return true
			}
		} else {
			if strings.HasSuffix(tmpPath, "/") {
				tmpPath = gstr.SubStr(tmpPath, 0, len(tmpPath)-1)
			}
			if urlPath == tmpPath {
				return true
			}
		}
	}
	return false
}

func GetUserKey(g *gin.Context) string {
	if userKey, exists := g.Get(KeyUserKey); exists {
		if key, ok := userKey.(string); ok {
			return key
		}
	}
	return ""
}

func GetRequestToken(g *gin.Context) (string, error) {
	authHeader := g.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			return "", errcode.InvalidParams.Err("Bearer param invalid")
		} else if parts[1] == "" {
			return "", errcode.InvalidParams.Err("Bearer param empty")
		}
		return parts[1], nil
	}
	authHeader = g.Query(KeyToken)
	if authHeader == "" {
		return "", errcode.InvalidParams.Err("token empty")
	}
	return authHeader, nil
}
