//nolint:wrapcheck  // we use echo-style handler returns, i.e. return c.JSON(..)
package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/adwski/vidi/internal/api/middleware"
	common "github.com/adwski/vidi/internal/api/model"
	"github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/generators"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Store interface {
	Create(ctx context.Context, u *model.User) error
	Authorize(ctx context.Context, u *model.User) error
}

type Service struct {
	*echo.Echo
	logger         *zap.Logger
	idGen          *generators.ID
	s              Store
	watchURLPrefix string
}

type ServiceConfig struct {
	APIPrefix      string
	WatchURLPrefix string
}

func NewService(cfg *ServiceConfig) *Service {
	e := middleware.GetEchoWithDefaultMiddleware()
	apiPrefix := fmt.Sprintf("%s/", strings.TrimRight(cfg.APIPrefix, "/"))

	svc := &Service{
		idGen:          generators.NewID(),
		watchURLPrefix: strings.TrimRight(cfg.WatchURLPrefix, "/"),
	}
	e.POST(apiPrefix+"register", svc.register)
	e.POST(apiPrefix+"login", svc.login)

	svc.Echo = e
	return svc
}

func (svc *Service) register(c echo.Context) error {
	var req model.UserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, common.Response{
			Error: "invalid auth params",
		})
	}
	id, err := svc.idGen.Get()
	if err != nil {
		svc.logger.Error("cannot generate user id", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, &common.Response{
			Error: common.InternalError,
		})
	}

	err = svc.s.Create(c.Request().Context(), model.NewUserFromRequest(id, &req))
	if err == nil {
		// TODO set jwt cookie
		return c.JSON(http.StatusInternalServerError, &common.Response{
			Message: "registration complete",
		})
	}

	switch {
	case errors.Is(err, model.ErrAlreadyExists):
		return c.JSON(http.StatusConflict, &common.Response{
			Error: err.Error(),
		})
	default:
		svc.logger.Error("internal error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, &common.Response{
			Error: common.InternalError,
		})
	}
}

func (svc *Service) login(c echo.Context) error {
	var req model.UserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, common.Response{
			Error: "invalid auth params",
		})
	}

	err := svc.s.Authorize(c.Request().Context(), model.NewUserFromRequest("", &req))
	if err == nil {
		// TODO set jwt cookie
		return c.JSON(http.StatusInternalServerError, &common.Response{
			Message: "login ok",
		})
	}

	switch {
	case errors.Is(err, model.ErrNotFound):
		return c.JSON(http.StatusNotFound, &common.Response{
			Error: err.Error(),
		})
	default:
		svc.logger.Error("internal error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, &common.Response{
			Error: common.InternalError,
		})
	}
}
