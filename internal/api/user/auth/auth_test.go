//nolint:lll // jwt tokens in test cases
package auth

import (
	"testing"
	"time"

	"github.com/labstack/gommon/random"

	"github.com/adwski/vidi/internal/api/user/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth_NewTokenForUser(t *testing.T) {
	a, err := NewAuth(&Config{
		Secret:       "superSecret",
		Domain:       "domain.com",
		SecureCookie: false,
		Expiration:   time.Hour,
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
		Secret:       "superSecret",
		Domain:       "domain.com",
		SecureCookie: false,
		Expiration:   time.Hour,
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
		Secret:       "superSecret",
		Domain:       "domain.com",
		SecureCookie: false,
		Expiration:   time.Hour,
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
		Secret:       "superSecret",
		Domain:       "domain.com",
		SecureCookie: false,
		Expiration:   time.Hour,
	})
	require.NoError(t, err)

	echoMW := a.EchoAuthServiceSide()
	var echoMWType echo.MiddlewareFunc
	assert.IsType(t, echoMWType, echoMW)
}

func TestAuth_MiddlewareUser(t *testing.T) {
	a, err := NewAuth(&Config{
		Secret:       "superSecret",
		Domain:       "domain.com",
		SecureCookie: false,
		Expiration:   time.Hour,
	})
	require.NoError(t, err)

	echoMW := a.EchoAuthUserSide()
	var echoMWType echo.MiddlewareFunc
	assert.IsType(t, echoMWType, echoMW)
}

func TestAuth_New(t *testing.T) {
	a, err := NewAuth(&Config{Secret: random.String(minSecretLen - 1)})
	require.Error(t, err)
	require.Nil(t, a)

	a, err = NewAuth(&Config{Secret: random.String(minSecretLen)})
	require.NoError(t, err)
	require.NotNil(t, a)
}

func TestAuth_parseToken(t *testing.T) {
	type args struct {
		token string
	}
	type want struct {
		claims *Claims
		errMsg string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid token",
			args: args{
				token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiJ0ZXN0dXNlcmlkIiwibmFtZSI6InRlc3RVc2VyIiwiZXhwIjo0NzE1ODM5MzY0fQ.QpWCWosIauu56J4LsEvpxBImwvl_mUoUMVyEeSRKL-M",
			},
			want: want{
				claims: &Claims{
					UserID: "testuserid",
					Name:   "testUser",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: &jwt.NumericDate{Time: time.Unix(4715839364, 0)},
					},
				},
			},
		},
		{
			name: "expired token",
			args: args{
				token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiJ0ZXN0dXNlcmlkIiwibmFtZSI6InRlc3RVc2VyIiwiZXhwIjoxNzE1ODM5MzY0fQ.YF7N1-AooC5naxX9RygOsN-6UqwjS52lPLtA_amSk5k",
			},
			want: want{
				errMsg: "cannot parse token: token has invalid claims: token is expired",
			},
		},
		{
			name: "empty userid token",
			args: args{
				token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdFVzZXIiLCJleHAiOjQ3MTU4MzkzNjR9.-xk_69_qo3_PY0_5-pD86q4hB3Ktp2Q3ivLrjYlN1xA",
			},
			want: want{
				errMsg: "user id is empty",
			},
		},
		{
			name: "no expiration time",
			args: args{
				token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdFVzZXIifQ.xSfUPYDMFUoZFQAOnR6wy1vO-k1jJExhcbiJc5eIfgw",
			},
			want: want{
				errMsg: "expiration claim is missing",
			},
		},
		{
			name: "invalid token",
			args: args{
				token: "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTcxNTg0MTQ1OSwiZXhwIjoxNzE1ODQ1MDU5fQ",
			},
			want: want{
				errMsg: "cannot parse token: token is malformed: token contains an invalid number of segments",
			},
		},
		{
			name: "invalid algorithm",
			args: args{
				token: "eyJ0eXAiOiJKV1QiLCJhbGciOiJFUzUxMiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTcxNTg0MTQ1OSwiZXhwIjoxNzE1ODQ1MDU5fQ.Ab_S-6Ec6o1JKSy91PIurPMABMD4CeQjG1Js75mfw34yOkutoL8BptpNJ2w9SnWZGl_Jlbt3JAsrvlj4oG5vXpYAAJRg-EOOX14NM_rdFlUfYrJAEtbE5TPK81lcGPQSiPwFOVASsgIiFZIrT8szznTXJsINM-8AS08Qh3kLtuhQOOZ6",
			},
			want: want{
				errMsg: "cannot parse token: token is unverifiable: error while executing keyfunc: unexpected signing method: ES512",
			},
		},
		{
			name: "no claims",
			args: args{
				token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.hD10zcHw1CBGmgjuxh4tkVZ9SDMLQE-zNDX7xlcYObo",
			},
			want: want{
				errMsg: "expiration claim is missing",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewAuth(&Config{Secret: "superSecret"})
			require.NoError(t, err)
			require.NotNil(t, a)

			claims, err := a.parseToken(tt.args.token)
			if tt.want.errMsg != "" {
				assert.EqualError(t, err, tt.want.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.claims, claims)
			}
		})
	}
}
