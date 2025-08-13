package auth

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ishaqcherry9/depend/pkg/errcode"
	"github.com/ishaqcherry9/depend/pkg/gin/response"
	"github.com/ishaqcherry9/depend/pkg/jwt"
)

type SigningMethodHMAC = jwt.SigningMethodHMAC
type Claims = jwt.Claims

var (
	HS256 = jwt.HS256
	HS384 = jwt.HS384
	HS512 = jwt.HS512
)

var (
	customSigningKey    []byte
	customSigningMethod *jwt.SigningMethodHMAC
	customExpire        time.Duration
	customIssuer        string

	errOption = errors.New("jwt option is nil, please initialize first, call middleware.InitAuth()")
)

type initAuthOptions struct {
	issuer        string
	signingMethod *SigningMethodHMAC
}

func defaultInitAuthOptions() *initAuthOptions {
	return &initAuthOptions{
		signingMethod: HS256,
	}
}

type InitAuthOption func(*initAuthOptions)

func (o *initAuthOptions) apply(opts ...InitAuthOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithInitAuthSigningMethod(sm *jwt.SigningMethodHMAC) InitAuthOption {
	return func(o *initAuthOptions) {
		o.signingMethod = sm
	}
}

func WithInitAuthIssuer(issuer string) InitAuthOption {
	return func(o *initAuthOptions) {
		o.issuer = issuer
	}
}

func InitAuth(signingKey []byte, expire time.Duration, opts ...InitAuthOption) {
	o := defaultInitAuthOptions()
	o.apply(opts...)

	customSigningKey = signingKey
	customExpire = expire
	customSigningMethod = o.signingMethod
	customIssuer = o.issuer
}

type GenerateTokenOption func(*generateTokenOptions)

type generateTokenOptions struct {
	fields map[string]interface{}
}

func (o *generateTokenOptions) apply(opts ...GenerateTokenOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithGenerateTokenFields(fields map[string]interface{}) GenerateTokenOption {
	return func(o *generateTokenOptions) {
		o.fields = fields
	}
}

func GenerateToken(uid string, opts ...GenerateTokenOption) (string, error) {
	if customSigningMethod == nil || len(customSigningKey) == 0 {
		panic(errOption)
	}

	genOpts := []jwt.GenerateTokenOption{
		jwt.WithGenerateTokenSignKey(customSigningKey),
		jwt.WithGenerateTokenSignMethod(customSigningMethod),
	}
	o := &generateTokenOptions{}
	o.apply(opts...)
	if len(o.fields) > 0 {
		genOpts = append(genOpts, jwt.WithGenerateTokenFields(o.fields))
	}

	claimsOpts := []jwt.RegisteredClaimsOption{
		jwt.WithExpires(customExpire),
	}
	if customIssuer != "" {
		claimsOpts = append(claimsOpts, jwt.WithIssuer(customIssuer))
	}
	genOpts = append(genOpts, jwt.WithGenerateTokenClaims(claimsOpts...))

	_, token, err := jwt.GenerateToken(uid, genOpts...)
	return token, err
}

func ParseToken(token string) (*jwt.Claims, error) {
	if customSigningMethod == nil {
		panic(errOption)
	}

	return jwt.ValidateToken(token, jwt.WithValidateTokenSignKey(customSigningKey))
}

func RefreshToken(claims *jwt.Claims) (string, error) {
	return claims.NewToken(customExpire, customSigningMethod, customSigningKey)
}

const HeaderAuthorizationKey = "Authorization"

type ExtraVerifyFn = func(claims *jwt.Claims, c *gin.Context) error

type AuthOption func(*authOptions)

type authOptions struct {
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

		claims, err := ParseToken(tokenString)
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
