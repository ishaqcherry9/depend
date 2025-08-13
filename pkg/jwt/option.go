package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ishaqcherry9/depend/pkg/krand"
)

type SigningMethodHMAC = jwt.SigningMethodHMAC

var (
	HS256 = jwt.SigningMethodHS256
	HS384 = jwt.SigningMethodHS384
	HS512 = jwt.SigningMethodHS512
)

var (
	defaultSigningKey    = []byte("CaqGzKLUsmWWbWI6F5EZbLwHsQeJ5RLyYTwBqa3mDKY6")
	defaultSigningMethod = HS256
	defaultExpire        = 24 * time.Hour
)

var (
	ErrTokenExpired = jwt.ErrTokenExpired
	errClaims       = errors.New("claims is not match")
	errNotMatch     = errors.New(" access token and refresh token is not match")
)

type registeredClaimsOptions struct {
	registeredClaims jwt.RegisteredClaims
}

func defaultRegisteredClaimsOptions(expire time.Duration, id string) *registeredClaimsOptions {
	now := time.Now()
	return &registeredClaimsOptions{
		registeredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        id,
		},
	}
}

type RegisteredClaimsOption func(*registeredClaimsOptions)

func (o *registeredClaimsOptions) apply(opts ...RegisteredClaimsOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithIssuer(issuer string) RegisteredClaimsOption {
	return func(o *registeredClaimsOptions) {
		o.registeredClaims.Issuer = issuer
	}
}

func WithSubject(subject string) RegisteredClaimsOption {
	return func(o *registeredClaimsOptions) {
		o.registeredClaims.Subject = subject
	}
}

func WithAudience(audience ...string) RegisteredClaimsOption {
	return func(o *registeredClaimsOptions) {
		o.registeredClaims.Audience = audience
	}
}

func WithExpires(d time.Duration) RegisteredClaimsOption {
	return func(o *registeredClaimsOptions) {
		o.registeredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(d))
	}
}

func WithDeadline(expiresAt time.Time) RegisteredClaimsOption {
	return func(o *registeredClaimsOptions) {
		o.registeredClaims.ExpiresAt = jwt.NewNumericDate(expiresAt)
	}
}

func WithNotBefore(notBefore time.Time) RegisteredClaimsOption {
	return func(o *registeredClaimsOptions) {
		o.registeredClaims.NotBefore = jwt.NewNumericDate(notBefore)
	}
}

func WithIssuedAt(issuedAt time.Time) RegisteredClaimsOption {
	return func(o *registeredClaimsOptions) {
		o.registeredClaims.IssuedAt = jwt.NewNumericDate(issuedAt)
	}
}

func WithJwtID(id string) RegisteredClaimsOption {
	return func(o *registeredClaimsOptions) {
		if id == "" {
			return
		}
		o.registeredClaims.ID = id
	}
}

type generateTokenOptions struct {
	signKey    []byte
	signMethod jwt.SigningMethod

	fields map[string]interface{}

	tokenClaimsOptions *registeredClaimsOptions
}

func defaultGenerateTokenOptions() *generateTokenOptions {
	return &generateTokenOptions{
		tokenClaimsOptions: defaultRegisteredClaimsOptions(defaultExpire, krand.NewStringID()),
		signKey:            defaultSigningKey,
		signMethod:         defaultSigningMethod,
	}
}

type GenerateTokenOption func(*generateTokenOptions)

func (o *generateTokenOptions) apply(opts ...GenerateTokenOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithGenerateTokenSignMethod(sm jwt.SigningMethod) GenerateTokenOption {
	return func(o *generateTokenOptions) {
		o.signMethod = sm
	}
}

func WithGenerateTokenSignKey(key []byte) GenerateTokenOption {
	return func(o *generateTokenOptions) {
		o.signKey = key
	}
}

func WithGenerateTokenFields(fields map[string]interface{}) GenerateTokenOption {
	return func(o *generateTokenOptions) {
		o.fields = fields
	}
}

func WithGenerateTokenClaims(opts ...RegisteredClaimsOption) GenerateTokenOption {
	return func(o *generateTokenOptions) {
		o.tokenClaimsOptions.apply(opts...)
	}
}

type validateTokenOptions struct {
	signKey []byte
}

func defaultValidateTokenOptions() *validateTokenOptions {
	return &validateTokenOptions{
		signKey: defaultSigningKey,
	}
}

type ValidateTokenOption func(*validateTokenOptions)

func (o *validateTokenOptions) apply(opts ...ValidateTokenOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithValidateTokenSignKey(key []byte) ValidateTokenOption {
	return func(o *validateTokenOptions) {
		if len(key) == 0 {
			return
		}
		o.signKey = key
	}
}

type refreshTokenOptions struct {
	signKey []byte
	expire  time.Duration
}

func defaultRefreshTokenOptions() *refreshTokenOptions {
	return &refreshTokenOptions{
		signKey: defaultSigningKey,
		expire:  defaultExpire,
	}
}

type RefreshTokenOption func(*refreshTokenOptions)

func (o *refreshTokenOptions) apply(opts ...RefreshTokenOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithRefreshTokenSignKey(key []byte) RefreshTokenOption {
	return func(o *refreshTokenOptions) {
		o.signKey = key
	}
}

func WithRefreshTokenExpire(expire time.Duration) RefreshTokenOption {
	return func(o *refreshTokenOptions) {
		o.expire = expire
	}
}

type generateTwoTokensOptions struct {
	signMethod jwt.SigningMethod
	signKey    []byte

	fields map[string]interface{}

	accessTokenClaimsOptions  *registeredClaimsOptions
	refreshTokenClaimsOptions *registeredClaimsOptions
}

func defaultGenerateTwoTokensOptions() *generateTwoTokensOptions {
	id := krand.NewStringID()
	return &generateTwoTokensOptions{
		accessTokenClaimsOptions:  defaultRegisteredClaimsOptions(time.Minute*30, id),
		refreshTokenClaimsOptions: defaultRegisteredClaimsOptions(time.Hour*24*30, id),

		signKey:    defaultSigningKey,
		signMethod: defaultSigningMethod,
	}
}

type GenerateTwoTokensOption func(*generateTwoTokensOptions)

func (o *generateTwoTokensOptions) apply(opts ...GenerateTwoTokensOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithGenerateTwoTokensSignMethod(sm jwt.SigningMethod) GenerateTwoTokensOption {
	return func(o *generateTwoTokensOptions) {
		o.signMethod = sm
	}
}

func WithGenerateTwoTokensSignKey(key []byte) GenerateTwoTokensOption {
	return func(o *generateTwoTokensOptions) {
		o.signKey = key
	}
}

func WithGenerateTwoTokensFields(fields map[string]interface{}) GenerateTwoTokensOption {
	return func(o *generateTwoTokensOptions) {
		o.fields = fields
	}
}

func WithGenerateTwoTokensAccessTokenClaims(opts ...RegisteredClaimsOption) GenerateTwoTokensOption {
	return func(o *generateTwoTokensOptions) {
		o.accessTokenClaimsOptions.apply(opts...)
	}
}

func WithGenerateTwoTokensRefreshTokenClaims(opts ...RegisteredClaimsOption) GenerateTwoTokensOption {
	return func(o *generateTwoTokensOptions) {
		o.refreshTokenClaimsOptions.apply(opts...)
	}
}

type refreshTwoTokensOptions struct {
	signKey            []byte
	accessTokenExpire  time.Duration
	refreshTokenExpire time.Duration
}

func defaultRefreshTwoTokensOptions() *refreshTwoTokensOptions {
	return &refreshTwoTokensOptions{
		signKey:            defaultSigningKey,
		accessTokenExpire:  time.Minute * 30,
		refreshTokenExpire: time.Hour * 24 * 30,
	}
}

type RefreshTwoTokensOption func(*refreshTwoTokensOptions)

func (o *refreshTwoTokensOptions) apply(opts ...RefreshTwoTokensOption) {
	for _, opt := range opts {
		opt(o)
	}
}

func WithRefreshTwoTokensSignKey(key []byte) RefreshTwoTokensOption {
	return func(o *refreshTwoTokensOptions) {
		o.signKey = key
	}
}

func WithRefreshTwoTokensRefreshTokenExpires(d time.Duration) RefreshTwoTokensOption {
	return func(o *refreshTwoTokensOptions) {
		o.refreshTokenExpire = d
	}
}

func WithRefreshTwoTokensAccessTokenExpires(d time.Duration) RefreshTwoTokensOption {
	return func(o *refreshTwoTokensOptions) {
		o.accessTokenExpire = d
	}
}

func getAlg(alg string) (jwt.SigningMethod, error) {
	switch alg {
	case "HS256":
		return HS256, nil
	case "HS384":
		return HS384, nil
	case "HS512":
		return HS512, nil
	default:
		return nil, errors.New("unsupported signing method: " + alg)
	}
}
