//nolint:wrapcheck  // we use echo-style handler returns, i.e. return c.JSON(..)
package video

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/adwski/vidi/internal/api/middleware"
	common "github.com/adwski/vidi/internal/api/model"
	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/generators"
	sessionStore "github.com/adwski/vidi/internal/session/store"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Store interface {
	Create(ctx context.Context, vi *model.Video) error
	Get(ctx context.Context, id string, userID string) (*model.Video, error)
	Delete(ctx context.Context, id string, userID string) error

	GetListByStatus(ctx context.Context, status model.Status) ([]*model.Video, error)
	Update(ctx context.Context, vi *model.Video) error
	UpdateLocation(ctx context.Context, vi *model.Video) error
	UpdateStatus(ctx context.Context, vi *model.Video) error
}

const (
	sessionTypeWatch  = "watch"
	sessionTypeUpload = "upload"

	uploadSessionDefaultTTL = 300 * time.Second
	watchSessionDefaultTTL  = 300 * time.Second
)

type Service struct {
	*echo.Echo
	logger          *zap.Logger
	idGen           *generators.ID
	auth            *auth.Auth
	watchSessions   *sessionStore.Store
	uploadSessions  *sessionStore.Store
	s               Store
	watchURLPrefix  string
	uploadURLPrefix string
}

type ServiceConfig struct {
	Logger          *zap.Logger
	Store           Store
	APIPrefix       string
	WatchURLPrefix  string
	UploadURLPrefix string
	RedisDSN        string
	AuthConfig      auth.Config
}

func NewService(cfg *ServiceConfig) (*Service, error) {
	var (
		apiPrefix          = strings.TrimRight(cfg.APIPrefix, "/")
		authenticator, err = auth.NewAuth(&cfg.AuthConfig)
	)
	if err != nil {
		return nil, fmt.Errorf("cannot configure authenticator: %w", err)
	}

	rUpload, errReU := sessionStore.NewStore(&sessionStore.Config{
		Name:     sessionTypeUpload,
		RedisDSN: cfg.RedisDSN,
		TTL:      uploadSessionDefaultTTL,
	})
	if errReU != nil {
		return nil, fmt.Errorf("cannot configure upload session store: %w", errReU)
	}

	rWatch, errReW := sessionStore.NewStore(&sessionStore.Config{
		Name:     sessionTypeWatch,
		RedisDSN: cfg.RedisDSN,
		TTL:      watchSessionDefaultTTL,
	})
	if errReW != nil {
		return nil, fmt.Errorf("cannot configure watch session store: %w", errReW)
	}

	svc := &Service{
		logger:          cfg.Logger,
		uploadSessions:  rUpload,
		watchSessions:   rWatch,
		auth:            authenticator,
		idGen:           generators.NewID(),
		watchURLPrefix:  strings.TrimRight(cfg.WatchURLPrefix, "/"),
		uploadURLPrefix: strings.TrimRight(cfg.UploadURLPrefix, "/"),
	}

	e := middleware.GetEchoWithDefaultMiddleware()
	// User zone
	api := e.Group(apiPrefix) // /api/video

	userAPI := api.Group("/user")
	userAPI.Use(authenticator.MiddlewareUser())
	userAPI.GET("/:id", svc.getVideo)
	userAPI.POST("/", svc.createVideo)
	userAPI.POST("/:id/watch", svc.watchVideo)
	userAPI.DELETE("/:id", svc.deleteVideo)

	// Service zone
	serviceAPI := api.Group("/service")
	serviceAPI.Use(authenticator.MiddlewareService())
	serviceAPI.PUT("/:id/location/:location", svc.updateVideoLocation)
	serviceAPI.PUT("/:id/status/:status", svc.updateVideoStatus)
	serviceAPI.PUT("/:id", svc.updateVideo)
	serviceAPI.GET("/search", svc.listVideos)

	svc.Echo = e
	return svc, nil
}

func (svc *Service) Handler() http.Handler {
	return svc.Echo
}

func (svc *Service) getUploadURL(sessionID string) string {
	return fmt.Sprintf("%s/%s", svc.uploadURLPrefix, sessionID)
}

func (svc *Service) getWatchURL(sessionID string) string {
	return fmt.Sprintf("%s/%s", svc.watchURLPrefix, sessionID)
}

func (svc *Service) commonResponse(c echo.Context, err error) error {
	if err == nil {
		return c.JSON(http.StatusOK, &common.Response{
			Message: "ok",
		})
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
