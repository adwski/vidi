//nolint:wrapcheck  // we use echo-style handler returns, i.e. return c.JSON(..)
package video

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/adwski/vidi/internal/mp4"
	"github.com/adwski/vidi/internal/session"

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

// Service is a Video API service. It has two "realms": user-side API and service-side API.
//
// User-side API provides CRUD operations with Video objects for a single user.
// While service-side API provides handlers for video updates by media processing services.
//
// Besides different API handlers they also differs in authentication approach:
// user-side only checks for valid user id in jwt cookie, while service-API
// looks up jwt in Bearer token and checks for valid service role.
//
// In production environment only user-side API should be exposed to public.
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
	Logger             *zap.Logger
	Store              Store
	APIPrefix          string
	WatchURLPrefix     string
	UploadURLPrefix    string
	UploadSessionStore *sessionStore.Store
	WatchSessionStore  *sessionStore.Store
	AuthConfig         auth.Config
}

func NewService(cfg *ServiceConfig) (*Service, error) {
	var (
		apiPrefix          = strings.TrimRight(cfg.APIPrefix, "/")
		authenticator, err = auth.NewAuth(&cfg.AuthConfig)
	)
	if err != nil {
		return nil, fmt.Errorf("cannot configure authenticator: %w", err)
	}

	svc := &Service{
		logger:          cfg.Logger,
		s:               cfg.Store,
		uploadSessions:  cfg.UploadSessionStore,
		watchSessions:   cfg.WatchSessionStore,
		auth:            authenticator,
		idGen:           generators.NewID(),
		watchURLPrefix:  strings.TrimRight(cfg.WatchURLPrefix, "/"),
		uploadURLPrefix: strings.TrimRight(cfg.UploadURLPrefix, "/"),
	}

	e := middleware.GetEchoWithDefaultMiddleware()
	api := e.Group(apiPrefix) // /api/video

	// User zone
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
	serviceAPI.POST("/search", svc.listVideos)

	svc.Echo = e
	return svc, nil
}

func (svc *Service) Handler() http.Handler {
	return svc.Echo
}

func (svc *Service) getUploadURL(sess *session.Session) string {
	return fmt.Sprintf("%s/%s", svc.uploadURLPrefix, sess.ID)
}

func (svc *Service) getWatchURL(sess *session.Session) string {
	return fmt.Sprintf("%s/%s/%s", svc.watchURLPrefix, sess.ID, mp4.MPDSuffix)
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
