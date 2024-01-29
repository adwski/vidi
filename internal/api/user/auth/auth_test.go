package auth

import (
	"testing"
	"time"

	"github.com/adwski/vidi/internal/api/user/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth_NewTokenForUser(t *testing.T) {
	a, err := NewAuth(&Config{
		Secret:     "superSecret",
		Domain:     "domain.com",
		HTTPS:      false,
		Expiration: time.Hour,
	})
	require.NoError(t, err)

	token, err := a.NewTokenForUser(&model.User{
		ID:   "qweqweqwe",
		Name: "name",
	})
	require.NoError(t, err)

	jt, err := jwt.ParseWithClaims(token, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
		return a.secret, nil
	})
	require.NoError(t, err)

	claims, ok := jt.Claims.(*Claims)
	require.True(t, ok, "token must have claims")

	assert.Equal(t, "qweqweqwe", claims.UserID)
	assert.Equal(t, "name", claims.Name)
}

func TestAuth_NewTokenForService(t *testing.T) {
	a, err := NewAuth(&Config{
		Secret:     "superSecret",
		Domain:     "domain.com",
		HTTPS:      false,
		Expiration: time.Hour,
	})
	require.NoError(t, err)

	token, err := a.NewTokenForService("svc1")
	require.NoError(t, err)

	jt, err := jwt.ParseWithClaims(token, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
		return a.secret, nil
	})
	require.NoError(t, err)

	claims, ok := jt.Claims.(*Claims)
	require.True(t, ok, "token must have claims")

	assert.Equal(t, "svc1", claims.Name)
	assert.True(t, claims.IsService())
	assert.True(t, claims.Role == roleNameService)
}

func TestAuth_CookieForUser(t *testing.T) {
	a, err := NewAuth(&Config{
		Secret:     "superSecret",
		Domain:     "domain.com",
		HTTPS:      false,
		Expiration: time.Hour,
	})
	require.NoError(t, err)

	cookie, err := a.CookieForUser(&model.User{
		ID:   "qweqweqwe",
		Name: "name",
	})
	require.NoError(t, err)

	assert.Equal(t, jwtCookieName, cookie.Name)
	assert.Equal(t, "domain.com", cookie.Domain)
	assert.False(t, cookie.Secure)
	assert.WithinDuration(t, time.Now().Add(time.Hour), cookie.Expires, time.Second)

	jt, err := jwt.ParseWithClaims(cookie.Value, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
		return a.secret, nil
	})
	require.NoError(t, err)
	require.True(t, jt.Valid)

	claims, ok := jt.Claims.(*Claims)
	require.True(t, ok, "token must have claims")

	assert.Equal(t, "qweqweqwe", claims.UserID)
	assert.Equal(t, "name", claims.Name)
}

func TestAuth_MiddlewareService(t *testing.T) {
	a, err := NewAuth(&Config{
		Secret:     "superSecret",
		Domain:     "domain.com",
		HTTPS:      false,
		Expiration: time.Hour,
	})
	require.NoError(t, err)

	echoMW := a.MiddlewareService()
	var echoMWType echo.MiddlewareFunc
	assert.IsType(t, echoMWType, echoMW)
}

func TestAuth_MiddlewareUser(t *testing.T) {
	a, err := NewAuth(&Config{
		Secret:     "superSecret",
		Domain:     "domain.com",
		HTTPS:      false,
		Expiration: time.Hour,
	})
	require.NoError(t, err)

	echoMW := a.MiddlewareUser()
	var echoMWType echo.MiddlewareFunc
	assert.IsType(t, echoMWType, echoMW)
}
