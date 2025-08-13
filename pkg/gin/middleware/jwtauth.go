package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/ishaqcherry9/depend/pkg/errcode"
	"github.com/ishaqcherry9/depend/pkg/gin/response"
	"github.com/ishaqcherry9/depend/pkg/jwt"
)

const HeaderAuthorizationKey = "Authorization"

type ExtraVerifyFn = func(claims *jwt.Claims, c *gin.Context) error

type AuthOption func(*authOptions)

type authOptions struct {
	signKey           []byte
	isReturnErrReason bool
	extraVerifyFn     ExtraVerifyFn
}

func defaultAuthOptions() *authOptions {
	return &authOptions{}
}

func (o *authOptions) apply(opts ...AuthOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithSignKey(key []byte) AuthOption {
	return func(o *authOptions) {
		o.signKey = key
	}
}

func WithReturnErrReason() AuthOption {
	return func(o *authOptions) {
		o.isReturnErrReason = true
	}
}

func WithExtraVerify(fn ExtraVerifyFn) AuthOption {
	return func(o *authOptions) {
		o.extraVerifyFn = fn
	}
}

var WithVerify = WithExtraVerify

func responseUnauthorized(isReturnErrReason bool, errMsg string) *errcode.Error {
	if isReturnErrReason {
		return errcode.Unauthorized.RewriteMsg("Unauthorized, " + errMsg)
	}
	return errcode.Unauthorized
}

func Auth(opts ...AuthOption) gin.HandlerFunc {
	o := defaultAuthOptions()
	o.apply(opts...)

	return func(c *gin.Context) {
		authorization := c.GetHeader(HeaderAuthorizationKey)
		if len(authorization) < 100 {
			response.Out(c, responseUnauthorized(o.isReturnErrReason, "token is illegal"))
			c.Abort()
			return
		}

		tokenString := authorization[7:]

		claims, err := jwt.ValidateToken(tokenString, jwt.WithValidateTokenSignKey(o.signKey))
		if err != nil {
			response.Out(c, responseUnauthorized(o.isReturnErrReason, err.Error()))
			c.Abort()
			return
		}

		if o.extraVerifyFn != nil {
			if err = o.extraVerifyFn(claims, c); err != nil {
				response.Out(c, responseUnauthorized(o.isReturnErrReason, err.Error()))
				c.Abort()
				return
			}
		}
		c.Set("claims", claims)
		c.Next()
	}
}

func GetClaims(c *gin.Context) (*jwt.Claims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}
	jwtClaims, ok := claims.(*jwt.Claims)
	return jwtClaims, ok
}
