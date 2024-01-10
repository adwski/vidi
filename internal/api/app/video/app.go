package video

import (
	"context"
	"net/http"

	"github.com/adwski/vidi/internal/api/app"
	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/adwski/vidi/internal/api/video"
	"github.com/adwski/vidi/internal/api/video/store"
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

func (a *App) configure(ctx context.Context) (http.Handler, app.Closer, bool) {
	var (
		logger = a.Logger()
		v      = a.Viper()
	)

	storeCfg := &store.Config{
		Logger: logger,
		DSN:    v.GetString("database.dsn"),
	}
	if v.HasErrors() {
		for param, errP := range v.Errors() {
			logger.Error("database param error", zap.String("param", param), zap.Error(errP))
		}
		return nil, nil, false
	}
	videoStorage, errStore := store.New(ctx, storeCfg)
	if errStore != nil {
		logger.Error("could not configure api storage", zap.Error(errStore))
		return nil, nil, false
	}

	svcCfg := &video.ServiceConfig{
		Logger:          logger,
		Store:           videoStorage,
		APIPrefix:       v.GetStringAllowEmpty("api.prefix"),
		WatchURLPrefix:  v.GetString("media.url.watch"),
		UploadURLPrefix: v.GetString("media.url.upload"),
		RedisDSN:        v.GetString("redis.dsn"),
		AuthConfig: auth.Config{
			Secret:     v.GetString("auth.jwt.secret"),
			Expiration: v.GetDuration("auth.jwt.expiration"),
			Domain:     v.GetString("domain"),
			HTTPS:      v.GetBool("https.enable"),
		},
	}
	if v.HasErrors() {
		for param, errP := range v.Errors() {
			logger.Error("api param error", zap.String("param", param), zap.Error(errP))
		}
		return nil, nil, false
	}

	svc, errSvc := video.NewService(svcCfg)
	if errSvc != nil {
		logger.Error("could not configure api service", zap.Error(errSvc))
		return nil, nil, false
	}

	return svc.Handler(), videoStorage, true
}
