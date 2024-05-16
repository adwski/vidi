package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/generators"
	"github.com/golang-jwt/jwt/v5"
)

const (
	minSecretLen = 8

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
	secureCookie  bool
}

type Config struct {
	Secret       string
	Domain       string
	SecureCookie bool
	Expiration   time.Duration
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
		secureCookie:  cfg.SecureCookie,
	}, nil
}

type Claims struct {
	UserID string `json:"uid"`
	Name   string `json:"name"`
	Role   string `json:"role,omitempty"`
	jwt.RegisteredClaims
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

func (a *Auth) parseToken(signedToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(signedToken, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot parse token: %w", err)
	}
	if !token.Valid {
		return nil, errors.New("token is not valid")
	}
	claims, ok := token.Claims.(*Claims)
	switch {
	case !ok:
		return nil, errors.New("token does not have claims")
	case claims.ExpiresAt == nil:
		return nil, errors.New("expiration claim is missing")
	case claims.ExpiresAt.Before(time.Now()):
		return nil, errors.New("token expired")
	case claims.UserID == "":
		return nil, errors.New("user id is empty")
	default:
		return claims, nil
	}
}

func (a *Auth) makeSignedJwt(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(a.signingMethod, claims)
	signedToken, err := token.SignedString(a.secret)
	if err != nil {
		return "", fmt.Errorf("cannot sign token: %w", err)
	}
	return signedToken, nil
}

func (c *Claims) IsService() bool {
	return c.Role == roleNameService
}

func (a *Auth) expirationTime() time.Time {
	return time.Now().Add(a.expiration)
}
