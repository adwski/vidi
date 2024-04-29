package auth

import (
	"errors"
	"github.com/adwski/vidi/internal/api/user/model"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"net/http"
)

const (
	jwtCookieName = "vidiSessID"

	sessionContextKey = "vUser"
)

func (a *Auth) EchoAuthUserSide() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		ContinueOnIgnoredError: false,
		ContextKey:             sessionContextKey,
		SigningKey:             a.secret,
		SigningMethod:          echojwt.AlgorithmHS256,
		TokenLookup:            "cookie:" + jwtCookieName,
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(Claims)
		},
	})
}

func (a *Auth) EchoAuthServiceSide() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		ContinueOnIgnoredError: false,
		ContextKey:             sessionContextKey,
		SigningKey:             a.secret,
		SigningMethod:          echojwt.AlgorithmHS256,
		TokenLookup:            "header:Authorization:Bearer ", // space is important!
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(Claims)
		},
	})
}

func GetClaimFromEchoContext(c echo.Context) (*Claims, error) {
	token, ok := c.Get(sessionContextKey).(*jwt.Token)
	if !ok || token == nil {
		return nil, errors.New("cannot get jwt token from session context")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || claims == nil {
		return nil, errors.New("cannot get claims from jwt token")
	}
	return claims, nil
}

func (a *Auth) CookieForUser(user *model.User) (*http.Cookie, error) {
	token, err := a.NewTokenForUser(user)
	if err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:    jwtCookieName,
		Value:   token,
		Domain:  a.domain,
		Expires: a.expirationTime(),
		Secure:  a.https,
	}, nil
}
