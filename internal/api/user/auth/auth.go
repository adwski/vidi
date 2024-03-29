package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/generators"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

const (
	minSecretLen = 8

	jwtCookieName = "vidiSessID"

	sessionContextKey = "vUser"

	roleNameService = "service"
)

// Auth us an authenticator that can
// 1) Issue jwt token for user or service
// 2) Create echo jwt middleware for jwt validation.
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
	UserID string `json:"uid"`
	Name   string `json:"name"`
	Role   string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

func (c *Claims) IsService() bool {
	return c.Role == roleNameService
}

func (a *Auth) expirationTime() time.Time {
	return time.Now().Add(a.expiration)
}

func (a *Auth) NewTokenForUser(user *model.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(a.expirationTime()),
		},
	}
	return a.makeSignedJwt(claims)
}

func (a *Auth) NewTokenForService(name string) (string, error) {
	id, err := generators.NewID().Get()
	if err != nil {
		return "", fmt.Errorf("cannot generate id: %w", err)
	}
	claims := &Claims{
		UserID: fmt.Sprintf("svc-%s", id),
		Name:   name,
		Role:   roleNameService,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(a.expirationTime()),
		},
	}
	return a.makeSignedJwt(claims)
}

func (a *Auth) makeSignedJwt(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(a.signingMethod, claims)
	signedToken, err := token.SignedString(a.secret)
	if err != nil {
		return "", fmt.Errorf("cannot sign token: %w", err)
	}
	return signedToken, nil
}

func (a *Auth) MiddlewareUser() echo.MiddlewareFunc {
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

func (a *Auth) MiddlewareService() echo.MiddlewareFunc {
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

func GetClaimFromContext(c echo.Context) (*Claims, error) {
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
