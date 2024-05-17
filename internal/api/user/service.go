//nolint:wrapcheck  // we use echo-style handler returns, i.e. return c.JSON(..)
package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	apihttp "github.com/adwski/vidi/internal/api/http"
	common "github.com/adwski/vidi/internal/api/model"
	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/generators"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Store interface {
	Create(ctx context.Context, u *model.User) error
	Get(ctx context.Context, u *model.User) error
}

// Service is a user service that provides
// user registration in authentication.
type Service struct {
	*echo.Echo
	auth   *auth.Auth
	logger *zap.Logger
	idGen  *generators.ID
	s      Store
}

type ServiceConfig struct {
	Logger     *zap.Logger
	Store      Store
	APIPrefix  string
	AuthConfig auth.Config
}

func NewService(cfg *ServiceConfig) (*Service, error) {
	authenticator, err := auth.NewAuth(&cfg.AuthConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot configure authenticator: %w", err)
	}

	var (
		e         = apihttp.GetEchoWithDefaultMiddleware()
		apiPrefix = strings.TrimRight(cfg.APIPrefix, "/")
	)

	svc := &Service{
		s:      cfg.Store,
		logger: cfg.Logger,
		auth:   authenticator,
		idGen:  generators.NewID(),
	}

	api := e.Group(apiPrefix)
	api.POST("/register", svc.register)
	api.POST("/login", svc.login)

	e.Validator = NewRequestValidator(cfg.Logger)
	svc.Echo = e
	return svc, nil
}

func (svc *Service) register(c echo.Context) error {
	req, err := getUserRequest(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &common.Response{
			Error: err.Error(),
		})
	}
	id, errID := svc.idGen.Get()
	if errID != nil {
		svc.logger.Error("cannot generate user uid", zap.Error(errID))
		return c.JSON(http.StatusInternalServerError, common.ResponseInternalError)
	}

	user := model.NewUserFromRequest(id, req)
	if err = svc.s.Create(c.Request().Context(), user); err == nil {
		cookie, errC := svc.auth.CookieForUser(user)
		if errC != nil {
			svc.logger.Error("cannot create auth cookie", zap.Error(errC))
			return c.JSON(http.StatusInternalServerError, common.ResponseInternalError)
		}
		c.SetCookie(cookie)
		return c.JSON(http.StatusOK, common.ResponseRegistrationComplete)
	}

	switch {
	case errors.Is(err, model.ErrAlreadyExists):
		return c.JSON(http.StatusConflict, &common.Response{
			Error: err.Error(),
		})
	default:
		svc.logger.Error("internal error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, common.ResponseInternalError)
	}
}

func (svc *Service) login(c echo.Context) error {
	req, err := getUserRequest(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &common.Response{
			Error: err.Error(),
		})
	}

	u := model.NewUserFromRequest("", req)
	err = svc.s.Get(c.Request().Context(), u) // side effect: Get() will fill user ID
	if err == nil {
		cookie, errC := svc.auth.CookieForUser(u)
		if errC != nil {
			svc.logger.Error("cannot create auth cookie", zap.Error(errC))
			return c.JSON(http.StatusInternalServerError, common.ResponseInternalError)
		}
		c.SetCookie(cookie)
		return c.JSON(http.StatusOK, common.ResponseOK)
	}

	switch {
	case errors.Is(err, model.ErrNotFound),
		errors.Is(err, model.ErrIncorrectCredentials):
		return c.JSON(http.StatusUnauthorized, common.ResponseIncorrectCredentials)
	default:
		svc.logger.Error("internal error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, &common.ResponseInternalError)
	}
}

func getUserRequest(c echo.Context) (*model.UserRequest, error) {
	var req model.UserRequest
	if err := c.Bind(&req); err != nil {
		return nil, errors.New("invalid params")
	}
	if err := c.Validate(req); err != nil {
		return nil, err
	}
	return &req, nil
}
