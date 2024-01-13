package video

import (
	"context"

	"github.com/adwski/vidi/internal/api/server"
	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/adwski/vidi/internal/api/video"
	"github.com/adwski/vidi/internal/api/video/store"
	"github.com/adwski/vidi/internal/app"
	"go.uber.org/zap"
)

type App struct {
	*app.App
}

func NewApp() *App {
	a := &App{}
	a.App = app.New(a.configure)
	return a
}

func (a *App) configure(ctx context.Context) (app.Runner, app.Closer, bool) {
	var (
		logger = a.Logger()
		v      = a.Viper()
	)
	storeCfg := &store.Config{
		Logger: logger,
		DSN:    v.GetString("database.dsn"),
	}
	svcCfg := &video.ServiceConfig{
		Logger:          logger,
		APIPrefix:       v.GetURIPrefix("api.prefix"),
		WatchURLPrefix:  v.GetURL("media.url.watch"),
		UploadURLPrefix: v.GetURL("media.url.upload"),
		RedisDSN:        v.GetURL("redis.dsn"),
		AuthConfig: auth.Config{
			Secret:     v.GetString("auth.jwt.secret"),
			Expiration: v.GetDuration("auth.jwt.expiration"),
			Domain:     v.GetString("domain"),
			HTTPS:      v.GetBool("https.enable"),
		},
	}
	srvCfg := &server.Config{
		Logger:            logger,
		ListenAddress:     v.GetString("server.address"),
		ReadTimeout:       v.GetDuration("server.timeouts.read"),
		ReadHeaderTimeout: v.GetDuration("server.timeouts.readHeader"),
		WriteTimeout:      v.GetDuration("server.timeouts.write"),
		IdleTimeout:       v.GetDuration("server.timeouts.idle"),
	}
	if v.HasErrors() {
		for param, errP := range v.Errors() {
			logger.Error("configuration error", zap.String("param", param), zap.Error(errP))
		}
		return nil, nil, false
	}
	videoStorage, errStore := store.New(ctx, storeCfg)
	if errStore != nil {
		logger.Error("could not configure api storage", zap.Error(errStore))
		return nil, nil, false
	}
	svcCfg.Store = videoStorage
	svc, errSvc := video.NewService(svcCfg)
	if errSvc != nil {
		logger.Error("could not configure service", zap.Error(errSvc))
		return nil, nil, false
	}
	srv, errSrv := server.NewServer(srvCfg)
	if errSrv != nil {
		logger.Error("could not configure server", zap.Error(errSrv))
		return nil, nil, false
	}
	srv.SetHandler(svc)
	return srv, videoStorage, true
}
