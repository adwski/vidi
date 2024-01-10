package user

import (
	"context"
	"net/http"

	"github.com/adwski/vidi/internal/api/app"
	"github.com/adwski/vidi/internal/api/user"
	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/adwski/vidi/internal/api/user/store"
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

func (app *App) configure(ctx context.Context) (http.Handler, app.Closer, bool) {
	var (
		logger = app.Logger()
		v      = app.Viper()
	)

	storeCfg := &store.Config{
		Logger: logger,
		DSN:    v.GetString("database.dsn"),
		Salt:   v.GetString("database.salt"),
	}
	if v.HasErrors() {
		for param, errP := range v.Errors() {
			logger.Error("database param error", zap.String("param", param), zap.Error(errP))
		}
		return nil, nil, false
	}
	userStorage, errStore := store.New(ctx, storeCfg)
	if errStore != nil {
		logger.Error("could not configure api storage", zap.Error(errStore))
		return nil, nil, false
	}

	svcCfg := &user.ServiceConfig{
		Logger:    logger,
		Store:     userStorage,
		APIPrefix: v.GetStringAllowEmpty("api.prefix"),
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

	svc, errSvc := user.NewService(svcCfg)
	if errSvc != nil {
		logger.Error("could not configure api service", zap.Error(errSvc))
		return nil, nil, false
	}
	return svc.Handler(), userStorage, true
}
