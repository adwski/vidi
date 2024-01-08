package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/adwski/vidi/internal/api/user/model"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

const (
	minSecretLen = 8

	jwtCookieName = "vidiSessID"
)

type Auth struct {
	signingMethod jwt.SigningMethod
	domain        string
	secret        []byte
	expiration    time.Duration
	https         bool
}

type Config struct {
	Secret     string
	Domain     string
	HTTPS      bool
	Expiration time.Duration
}

func NewAuth(cfg *Config) (*Auth, error) {
	if len(cfg.Secret) < minSecretLen {
		return nil, fmt.Errorf("secret length cannot be less than %d", minSecretLen)
	}
	return &Auth{
		signingMethod: jwt.SigningMethodHS256,
		secret:        []byte(cfg.Secret),
		expiration:    cfg.Expiration,
		domain:        cfg.Domain,
		https:         cfg.HTTPS,
	}, nil
}

type Claims struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
	jwt.RegisteredClaims
}

func (a *Auth) expirationTime() time.Time {
	return time.Now().Add(a.expiration)
}

func (a *Auth) NewTokenForUser(user *model.User) (string, error) {
	claims := &Claims{
		UID:  user.UID,
		Name: user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(a.expirationTime()),
		},
	}
	token := jwt.NewWithClaims(a.signingMethod, claims)
	signedToken, err := token.SignedString(a.secret)
	if err != nil {
		return "", fmt.Errorf("cannot sign token: %w", err)
	}
	return signedToken, nil
}

func (a *Auth) Middleware() echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		ContinueOnIgnoredError: false,
		SigningKey:             a.secret,
		SigningMethod:          echojwt.AlgorithmHS256,
		TokenLookup:            "cookie:jwtCookieName",
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(Claims)
		},
	})
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
