//nolint:wrapcheck  // we use echo-style handler returns, i.e. return c.JSON(..)
package video

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/adwski/vidi/internal/api/middleware"
	common "github.com/adwski/vidi/internal/api/model"
	user "github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/generators"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Store interface {
	Get(ctx context.Context, id string) (*model.Video, error)
	Create(ctx context.Context, vi *model.Video) error
	Update(ctx context.Context, vi *model.Video) error
	Delete(ctx context.Context, id string) error
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
	e.GET(apiPrefix+":id", svc.getVideo)
	e.POST(apiPrefix, svc.createVideo)
	e.DELETE(apiPrefix+":id", svc.deleteVideo)
	e.POST(apiPrefix+":id/watch", svc.watchVideo)

	svc.Echo = e
	return svc
}

func (svc *Service) getPathURL(sessionID string) string {
	return fmt.Sprintf("%svc/%svc", svc.watchURLPrefix, sessionID)
}

func (svc *Service) watchVideo(c echo.Context) error {
	video, err := svc.s.Get(c.Request().Context(), c.Param("id"))
	if err != nil {
		return svc.erroredResponse(c, err)
	}

	if video.IsErrored() {
		return c.JSON(http.StatusOK, &common.Response{
			Error: "video cannot be watched",
		})
	}

	if !video.IsReady() {
		return c.JSON(http.StatusOK, &common.Response{
			Message: "video is not ready",
		})
	}

	// TODO Add watch session handling
	sessionID := "xxx"
	return c.JSON(http.StatusOK, &model.WatchResponse{
		WatchURL: svc.getPathURL(sessionID),
	})
}

func (svc *Service) deleteVideo(c echo.Context) error {
	err := svc.s.Delete(c.Request().Context(), c.Param("id"))
	if err == nil {
		return c.JSON(http.StatusOK, &common.Response{
			Message: "ok",
		})
	}
	return svc.erroredResponse(c, err)
}

func (svc *Service) createVideo(c echo.Context) error {
	newID, err := svc.idGen.Get()
	if err != nil {
		svc.logger.Error("cannot generate new video id", zap.Error(err))
		return c.JSON(http.StatusNotFound, &common.Response{
			Error: common.InternalError,
		})
	}

	newVideo := model.NewVideo(newID, &user.User{}) // TODO add user session handling
	err = svc.s.Create(c.Request().Context(), newVideo)

	url := "" // TODO add upload session handling

	if err == nil {
		return c.JSON(http.StatusOK, newVideo.UploadResponse(url))
	}
	switch {
	case errors.Is(err, model.ErrAlreadyExists):
		svc.logger.Error("video with generated id already exists")
	default:
		svc.logger.Error("cannot create video in storage", zap.Error(err))
	}
	return c.JSON(http.StatusInternalServerError, &common.Response{
		Error: common.InternalError,
	})
}

func (svc *Service) getVideo(c echo.Context) error {
	video, err := svc.s.Get(c.Request().Context(), c.Param("id"))
	if err == nil {
		return c.JSON(http.StatusOK, video.Response())
	}
	return svc.erroredResponse(c, err)
}

func (svc *Service) erroredResponse(c echo.Context, err error) error {
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
