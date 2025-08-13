package jwt

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	uid          = "123"
	customFields = map[string]interface{}{
		"name":       "john",
		"role":       "admin",
		"department": "engineering",
		"age":        11,
		"is_active":  true,
		"price":      3.14,
	}
)

func getRegisteredClaimsOptions(ds ...time.Duration) []RegisteredClaimsOption {
	now := time.Now()

	d := time.Hour * 2
	if len(ds) > 0 {
		d = ds[0]
	}
	return []RegisteredClaimsOption{
		WithIssuer("https://auth.example.com"),
		WithSubject("123"),
		WithAudience("https://api.example.com"),
		WithIssuedAt(now),
		WithNotBefore(now),
		WithExpires(d),
		WithDeadline(now.Add(d)),
		WithJwtID("abc1234xxx"),
	}
}

func TestGenerateAndValidateToken(t *testing.T) {

	t.Run("GenerateAndValidateToken, default options", func(t *testing.T) {
		jwtID, token, err := GenerateToken(uid)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotEmpty(t, jwtID)
		t.Log(token)

		claims, err := ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, uid, claims.UID)
	})

	t.Run("GenerateAndValidateToken, have options", func(t *testing.T) {
		signMethod := HS384
		signKey := []byte("your-secret-key")
		jwtID, token, err := GenerateToken(
			uid,
			WithGenerateTokenFields(customFields),
			WithGenerateTokenSignMethod(signMethod),
			WithGenerateTokenSignKey(signKey),
			WithGenerateTokenClaims(getRegisteredClaimsOptions()...),
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.NotEmpty(t, jwtID)
		t.Log(jwtID, token)

		claims, err := ValidateToken(token, WithValidateTokenSignKey(signKey))
		assert.NoError(t, err)
		assert.Equal(t, uid, claims.UID)
		assert.Equal(t, customFields["name"], claims.Fields["name"])
	})

	t.Run("GenerateAndValidateToken, error test", func(t *testing.T) {
		token2 := "xxx.xxx.xxx"
		_, err := ValidateToken(token2)
		assert.Error(t, err)
		_, token, _ := GenerateToken(uid)
		token3 := token + "xxx"
		_, err = ValidateToken(token3)
		assert.Error(t, err)

		_, token, err = GenerateToken(uid, WithGenerateTokenClaims(WithExpires(time.Millisecond*200)))
		assert.NoError(t, err)
		time.Sleep(time.Second)
		_, err = ValidateToken(token)
		assert.True(t, errors.Is(err, ErrTokenExpired))
	})
}

func TestRefreshToken(t *testing.T) {

	t.Run("RefreshToken, default options", func(t *testing.T) {
		jwtID, token, err := GenerateToken(uid)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		t.Log(jwtID, "\n", token)

		time.Sleep(time.Second)
		jwtID2, newToken, err := RefreshToken(token)
		assert.NoError(t, err)
		assert.NotEmpty(t, newToken)
		t.Log(jwtID, "\n", newToken)

		assert.Equal(t, jwtID, jwtID2)
		assert.NotEqual(t, token, newToken)
	})

	t.Run("RefreshToken, have options", func(t *testing.T) {
		signMethod := HS512
		signKey := []byte("your-secret-key")
		jwtID, token, err := GenerateToken(
			uid,
			WithGenerateTokenFields(customFields),
			WithGenerateTokenSignMethod(signMethod),
			WithGenerateTokenSignKey(signKey),
			WithGenerateTokenClaims(getRegisteredClaimsOptions()...),
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		t.Log(token)

		time.Sleep(time.Second)
		jwtID2, newToken, err := RefreshToken(
			token,
			WithRefreshTokenSignKey(signKey),
			WithRefreshTokenExpire(time.Hour),
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, newToken)
		t.Log(newToken)

		assert.Equal(t, jwtID, jwtID2)
		assert.NotEqual(t, token, newToken)
	})
}

func TestGenerateAndValidateTwoTokens(t *testing.T) {

	t.Run("GenerateAndValidateTwoTokens, default options", func(t *testing.T) {
		tokens, err := GenerateTwoTokens(uid)
		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		t.Log(tokens.JwtID, "\n", tokens.RefreshToken, "\n", tokens.AccessToken)

		refreshClaims, err := ValidateToken(tokens.RefreshToken)
		assert.NoError(t, err)
		assert.Equal(t, uid, refreshClaims.UID)
		assert.Equal(t, tokens.JwtID, refreshClaims.ID)

		accessClaims, err := ValidateToken(tokens.AccessToken)
		assert.NoError(t, err)
		assert.Equal(t, uid, accessClaims.UID)
		assert.Equal(t, tokens.JwtID, accessClaims.ID)
	})

	t.Run("GenerateAndValidateTwoTokens, custom options", func(t *testing.T) {
		signMethod := HS256
		signKey := []byte("your-secret-key")
		tokens, err := GenerateTwoTokens(
			uid,
			WithGenerateTwoTokensSignMethod(signMethod),
			WithGenerateTwoTokensSignKey(signKey),
			WithGenerateTwoTokensFields(customFields),
			WithGenerateTwoTokensRefreshTokenClaims(getRegisteredClaimsOptions(time.Hour*24)...),
			WithGenerateTwoTokensAccessTokenClaims(getRegisteredClaimsOptions()...),
		)
		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.NotEqual(t, tokens.RefreshToken, tokens.AccessToken)
		t.Log(tokens.JwtID, "\n", tokens.RefreshToken, "\n", tokens.AccessToken)

		refreshClaims, err := ValidateToken(tokens.RefreshToken, WithValidateTokenSignKey(signKey))
		assert.NoError(t, err)
		assert.Equal(t, uid, refreshClaims.UID)
		assert.Equal(t, customFields["name"], refreshClaims.Fields["name"])
		assert.Equal(t, tokens.JwtID, refreshClaims.ID)

		accessClaims, err := ValidateToken(tokens.AccessToken, WithValidateTokenSignKey(signKey))
		assert.NoError(t, err)
		assert.Equal(t, uid, accessClaims.UID)
		assert.Equal(t, customFields["name"], accessClaims.Fields["name"])
		assert.Equal(t, tokens.JwtID, accessClaims.ID)
	})
}

func TestRefreshTwoToken(t *testing.T) {

	t.Run("RefreshTwoToken, default options", func(t *testing.T) {
		tokens, err := GenerateTwoTokens(uid)
		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		t.Log(tokens.JwtID, "\n", tokens.RefreshToken, "\n", tokens.AccessToken)

		time.Sleep(time.Second)
		newTokens, err := RefreshTwoTokens(tokens.RefreshToken, tokens.AccessToken)
		assert.NoError(t, err)
		assert.NotNil(t, newTokens)
		assert.Equal(t, tokens.JwtID, newTokens.JwtID)
		t.Log(newTokens.JwtID, "\n", newTokens.RefreshToken, "\n", newTokens.AccessToken)

		assert.Equal(t, newTokens.RefreshToken, tokens.RefreshToken)
		assert.NotEqual(t, newTokens.AccessToken, tokens.AccessToken)
	})

	t.Run("RefreshTwoToken, have options", func(t *testing.T) {
		signMethod := HS384
		signKey := []byte("your-secret-key")
		tokens, err := GenerateTwoTokens(
			uid,
			WithGenerateTwoTokensFields(customFields),
			WithGenerateTwoTokensSignMethod(signMethod),
			WithGenerateTwoTokensSignKey(signKey),
			WithGenerateTwoTokensRefreshTokenClaims(getRegisteredClaimsOptions(time.Hour*24)...),
			WithGenerateTwoTokensAccessTokenClaims(getRegisteredClaimsOptions()...),
		)
		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		t.Log(tokens.JwtID, "\n", tokens.RefreshToken, "\n", tokens.AccessToken)
		assert.NotEqual(t, tokens.RefreshToken, tokens.AccessToken)

		time.Sleep(time.Second)
		newTokens, err := RefreshTwoTokens(
			tokens.RefreshToken,
			tokens.AccessToken,
			WithRefreshTwoTokensSignKey(signKey),
			WithRefreshTwoTokensRefreshTokenExpires(time.Hour*2),
			WithRefreshTwoTokensAccessTokenExpires(time.Minute*15),
		)
		assert.NoError(t, err)
		assert.NotNil(t, newTokens)
		assert.Equal(t, tokens.JwtID, newTokens.JwtID)
		t.Log(newTokens.JwtID, "\n", newTokens.RefreshToken, "\n", newTokens.AccessToken)

		refreshClaims, err := ValidateToken(tokens.RefreshToken, WithValidateTokenSignKey(signKey))
		assert.NoError(t, err)
		if refreshClaims.ExpiresAt.Sub(time.Now()) < time.Hour*3 {
			assert.NotEqual(t, newTokens.RefreshToken, tokens.RefreshToken)
		} else {
			assert.Equal(t, newTokens.RefreshToken, tokens.RefreshToken)
		}
		assert.NotEqual(t, newTokens.AccessToken, tokens.AccessToken)
	})
}

func TestClaims(t *testing.T) {
	_, token, _ := GenerateToken(
		uid,
		WithGenerateTokenFields(customFields),
		WithGenerateTokenClaims(getRegisteredClaimsOptions()...),
	)
	claims, _ := ValidateToken(token)

	name, _ := claims.GetString("name")
	assert.Equal(t, name, customFields["name"])

	age, _ := claims.GetInt("age")
	assert.Equal(t, age, customFields["age"])
	age2, _ := claims.GetInt64("age")
	assert.Equal(t, int(age2), customFields["age"])

	is_active, _ := claims.GetBool("is_active")
	assert.Equal(t, is_active, customFields["is_active"])

	price, _ := claims.GetFloat64("price")
	assert.Equal(t, price, customFields["price"])

	_, isExist := claims.Get("unknown_field")
	assert.Equal(t, false, isExist)
}

func TestGetClaimsUnverified(t *testing.T) {
	signMethod := HS512
	signKey := []byte("your-secret-key")
	_, tokenStr, err := GenerateToken(
		uid,
		WithGenerateTokenFields(customFields),
		WithGenerateTokenSignMethod(signMethod),
		WithGenerateTokenSignKey(signKey),
		WithGenerateTokenClaims(getRegisteredClaimsOptions(time.Second)...),
	)
	time.Sleep(time.Second * 2)
	claims, err := GetClaimsUnverified(tokenStr)
	assert.NoError(t, err)
	assert.Equal(t, claims.UID, "123")
	assert.Equal(t, claims.Fields["name"], "john")
}
