package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UID    string                 `json:"uid,omitempty"`
	Fields map[string]interface{} `json:"fields,omitempty"`
	jwt.RegisteredClaims
}

func (c *Claims) Get(key string) (val interface{}, isExist bool) {
	if c.Fields == nil {
		return nil, false
	}
	val, isExist = c.Fields[key]
	return val, isExist
}

func (c *Claims) GetString(key string) (string, bool) {
	val, isExist := c.Get(key)
	if isExist {
		str, ok := val.(string)
		return str, ok
	}
	return "", false
}

func (c *Claims) GetInt(key string) (int, bool) {
	val, isExist := c.Get(key)
	if isExist {
		if v, ok := val.(float64); ok {
			return int(v), true
		}
		if v, ok := val.(int); ok {
			return v, true
		}
	}
	return 0, false
}

func (c *Claims) GetInt64(key string) (uint64, bool) {
	val, isExist := c.Get(key)
	if isExist {
		if v, ok := val.(float64); ok {
			return uint64(v), true
		}
		if v, ok := val.(uint64); ok {
			return v, true
		}
	}
	return 0, false
}

func (c *Claims) GetBool(key string) (b bool, isExist bool) {
	val, exist := c.Get(key)
	if exist {
		v, ok := val.(bool)
		return v, ok
	}
	return false, false
}

func (c *Claims) GetFloat64(key string) (float64, bool) {
	val, isExist := c.Get(key)
	if isExist {
		v, ok := val.(float64)
		return v, ok
	}
	return 0, false
}

func (c *Claims) NewToken(d time.Duration, signMethod jwt.SigningMethod, signKey []byte) (string, error) {
	now := time.Now()
	c.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(now.Add(d))
	c.RegisteredClaims.IssuedAt = jwt.NewNumericDate(now)
	if c.RegisteredClaims.NotBefore != nil {
		c.RegisteredClaims.NotBefore = jwt.NewNumericDate(now)
	}
	token := jwt.NewWithClaims(signMethod, c)
	return token.SignedString(signKey)
}

func GetClaimsUnverified(tokenString string) (*Claims, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errClaims
	}
	return claims, nil
}

func GenerateToken(uid string, opts ...GenerateTokenOption) (jwtID string, tokenStr string, err error) {
	o := defaultGenerateTokenOptions()
	o.apply(opts...)

	claims := Claims{
		uid,
		o.fields,
		o.tokenClaimsOptions.registeredClaims,
	}
	token := jwt.NewWithClaims(o.signMethod, claims)
	tokenStr, err = token.SignedString(o.signKey)
	return o.tokenClaimsOptions.registeredClaims.ID, tokenStr, err
}

func ValidateToken(tokenString string, opts ...ValidateTokenOption) (*Claims, error) {
	_, claims, err := verifyToken(tokenString, opts...)
	return claims, err
}

func verifyToken(tokenString string, opts ...ValidateTokenOption) (string, *Claims, error) {
	o := defaultValidateTokenOptions()
	o.apply(opts...)

	alg := ""
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		alg = token.Header["alg"].(string)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", alg)
		}
		return o.signKey, nil
	})
	if err != nil {
		return "", nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", nil, errClaims
	}
	return alg, claims, nil
}

func RefreshToken(tokenString string, opts ...RefreshTokenOption) (jwtID string, tokenStr string, err error) {
	o := defaultRefreshTokenOptions()
	o.apply(opts...)

	alg, claims, err := verifyToken(tokenString, WithValidateTokenSignKey(o.signKey))
	if err != nil {
		return "", "", err
	}

	signMethod, err := getAlg(alg)
	if err != nil {
		return "", "", err
	}

	tokenStr, err = claims.NewToken(o.expire, signMethod, o.signKey)
	return claims.ID, tokenStr, err
}

type Tokens struct {
	RefreshToken string `json:"refreshToken"`
	AccessToken  string `json:"accessToken"`
	JwtID        string `json:"jwtID"`
}

func GenerateTwoTokens(uid string, opts ...GenerateTwoTokensOption) (*Tokens, error) {
	o := defaultGenerateTwoTokensOptions()
	o.apply(opts...)

	o.accessTokenClaimsOptions.registeredClaims.ID = o.refreshTokenClaimsOptions.registeredClaims.ID

	claims := Claims{uid, o.fields, o.accessTokenClaimsOptions.registeredClaims}
	accessToken := jwt.NewWithClaims(o.signMethod, claims)
	accessTokenStr, err := accessToken.SignedString(o.signKey)
	if err != nil {
		return nil, err
	}

	claims.RegisteredClaims = o.refreshTokenClaimsOptions.registeredClaims
	refreshToken := jwt.NewWithClaims(o.signMethod, claims)
	refreshTokenStr, err := refreshToken.SignedString(o.signKey)
	if err != nil {
		return nil, err
	}

	return &Tokens{
		RefreshToken: refreshTokenStr,
		AccessToken:  accessTokenStr,
		JwtID:        o.refreshTokenClaimsOptions.registeredClaims.ID,
	}, nil
}

func RefreshTwoTokens(refreshToken string, accessToken string, opts ...RefreshTwoTokensOption) (*Tokens, error) {
	o := defaultRefreshTwoTokensOptions()
	o.apply(opts...)

	alg, refreshTokenClaims, err := verifyToken(refreshToken, WithValidateTokenSignKey(o.signKey))
	if err != nil {
		return nil, err
	}

	accessTokenClaims, err := GetClaimsUnverified(accessToken)
	if err != nil {
		return nil, err
	}

	if refreshTokenClaims.ID != accessTokenClaims.ID || refreshTokenClaims.UID != accessTokenClaims.UID {
		return nil, errNotMatch
	}

	signMethod, err := getAlg(alg)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if refreshTokenClaims.ExpiresAt.Sub(now) < time.Hour*3 {
		refreshToken, err = refreshTokenClaims.NewToken(o.refreshTokenExpire, signMethod, o.signKey)
		if err != nil {
			return nil, err
		}
	}
	accessToken, err = accessTokenClaims.NewToken(o.accessTokenExpire, signMethod, o.signKey)
	if err != nil {
		return nil, err
	}

	return &Tokens{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		JwtID:        refreshTokenClaims.ID,
	}, nil
}
