package server

import (
	"errors"
	"fmt"
	apihttp "github.com/adwski/vidi/internal/api/http"
	"github.com/adwski/vidi/internal/api/http/server"
	common "github.com/adwski/vidi/internal/api/model"
	"github.com/adwski/vidi/internal/api/user/auth"
	user "github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/api/video"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type Server struct {
	*server.Server
	logger   *zap.Logger
	videoSvc *video.Service
}

type Config struct {
	Logger     *zap.Logger
	AuthConfig auth.Config
	APIPrefix  string
	HTTPConfig *server.Config
}

func NewServer(cfg *Config, videoSvc *video.Service) (*Server, error) {
	var (
		apiPrefix          = strings.TrimRight(cfg.APIPrefix, "/")
		authenticator, err = auth.NewAuth(&cfg.AuthConfig)
	)
	if err != nil {
		return nil, fmt.Errorf("cannot configure authenticator: %w", err)
	}

	// echo router with preconfigured middleware
	e := apihttp.GetEchoWithDefaultMiddleware()

	srv := &Server{
		logger:   cfg.Logger.With(zap.String("component", "http-api")),
		videoSvc: videoSvc,
	}

	// spawn server
	if srv.Server, err = server.NewServer(cfg.HTTPConfig); err != nil {
		return nil, fmt.Errorf("cannot configure http server: %w", err)
	}

	// Common api prefix, i.e. /api/video
	api := e.Group(apiPrefix)

	// User zone
	userAPI := api.Group("/user")
	userAPI.Use(authenticator.EchoAuthUserSide())
	userAPI.GET("/:id", srv.getVideo)
	userAPI.GET("/", srv.getVideos)
	userAPI.POST("/", srv.createVideo)
	userAPI.POST("/:id/watch", srv.watchVideo)
	userAPI.DELETE("/:id", srv.deleteVideo)
	quotaAPI := api.Group("/quota")
	quotaAPI.GET("/", srv.getQuota)

	// Service zone
	serviceAPI := api.Group("/service")
	serviceAPI.Use(authenticator.EchoAuthServiceSide())
	serviceAPI.PUT("/:id/status/:status", srv.updateVideoStatus)
	serviceAPI.PUT("/:id", srv.updateVideo)
	serviceAPI.POST("/search", srv.searchVideos)

	srv.SetHandler(e)
	return srv, nil
}

func (srv *Server) getUser(c echo.Context) (*user.User, error, bool) {
	claims, err := auth.GetClaimFromEchoContext(c)
	if err != nil {
		srv.logger.Debug("cannot get user from context", zap.Error(err))
		return nil, c.JSON(http.StatusUnauthorized, common.ResponseUnauthorized), false
	}
	return &user.User{
		ID:   claims.UserID,
		Name: claims.Name,
	}, nil, true
}

func (srv *Server) serviceAuth(c echo.Context) error {
	claims, err := auth.GetClaimFromEchoContext(c)
	if err != nil {
		srv.logger.Error("service auth fail", zap.Any("claims", claims))
		return c.JSON(http.StatusUnauthorized, common.ResponseUnauthorized)
	}
	logf := srv.logger.With(zap.String("id", claims.UserID),
		zap.String("name", claims.Name),
		zap.String("role", claims.Role))

	if claims.IsService() {
		logf.Debug("service auth ok")
		return nil
	}
	logf.Error("service auth incorrect role")
	return c.JSON(http.StatusUnauthorized, common.ResponseUnauthorized)
}

func (srv *Server) commonResponse(c echo.Context, err error) error {
	if err == nil {
		return c.JSON(http.StatusOK, common.ResponseOK)
	}
	return srv.erroredResponse(c, err)
}

func (srv *Server) erroredResponse(c echo.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrNotFound):
		return c.JSON(http.StatusNotFound, &common.Response{
			Error: err.Error(),
		})
	default:
		srv.logger.Error("internal error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, common.ResponseInternalError)
	}
}
