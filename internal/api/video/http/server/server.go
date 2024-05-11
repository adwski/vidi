//nolint:wrapcheck  // we use echo-style handler returns, i.e. return c.JSON(..)
package server

import (
	"errors"
	"fmt"
	"net/http"

	apihttp "github.com/adwski/vidi/internal/api/http"
	"github.com/adwski/vidi/internal/api/http/server"
	common "github.com/adwski/vidi/internal/api/model"
	"github.com/adwski/vidi/internal/api/user/auth"
	user "github.com/adwski/vidi/internal/api/user/model"
	"github.com/adwski/vidi/internal/api/video"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Server struct {
	*server.Server
	logger   *zap.Logger
	videoSvc *video.Service
}

type Config struct {
	Logger     *zap.Logger
	Auth       *auth.Auth
	HTTPConfig *server.Config
	APIPrefix  string
}

func NewServer(cfg *Config, videoSvc *video.Service) (*Server, error) {
	if cfg.Auth == nil {
		return nil, errors.New("authenticator cannot be nil")
	}

	// spawn server
	httpSrv, err := server.NewServer(cfg.HTTPConfig)
	if err != nil {
		return nil, fmt.Errorf("cannot configure http server: %w", err)
	}

	srv := &Server{
		logger:   cfg.Logger.With(zap.String("component", "http-api")),
		Server:   httpSrv,
		videoSvc: videoSvc,
	}

	// echo router with preconfigured middleware
	e := apihttp.GetEchoWithDefaultMiddleware()

	// Common api prefix, i.e. /api
	api := e.Group(cfg.APIPrefix)

	// Video zone
	videoAPI := api.Group("/video")
	videoAPI.Use(cfg.Auth.EchoAuthUserSide())
	videoAPI.GET("/:id", srv.getVideo)
	videoAPI.GET("/", srv.getVideos)
	videoAPI.POST("/", srv.createVideo)
	videoAPI.DELETE("/:id", srv.deleteVideo)

	// Watch zone
	watchAPI := api.Group("/watch")
	watchAPI.Use(cfg.Auth.EchoAuthUserSide())
	watchAPI.GET("/:id", srv.watchVideo)

	// Quota zone
	quotaAPI := api.Group("/quota")
	quotaAPI.Use(cfg.Auth.EchoAuthUserSide())
	quotaAPI.GET("/", srv.getQuota)

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

func (srv *Server) erroredResponse(c echo.Context, err error) error {
	switch {
	case errors.Is(err, model.ErrNotFound):
		return c.JSON(http.StatusNotFound, &common.Response{
			Error: err.Error(),
		})
	case errors.Is(err, model.ErrNotResumable):
		return c.JSON(http.StatusNotAcceptable, &common.Response{Error: err.Error()})
	default:
		srv.logger.Error("internal error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, common.ResponseInternalError)
	}
}
