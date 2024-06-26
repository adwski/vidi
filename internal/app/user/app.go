package user

import (
	"context"

	"github.com/adwski/vidi/internal/api/http/server"
	"github.com/adwski/vidi/internal/api/user"
	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/adwski/vidi/internal/api/user/store"
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

func (a *App) configure(ctx context.Context) ([]app.Runner, []app.Closer, bool) {
	var (
		logger = a.Logger()
		v      = a.Viper()
	)

	storeCfg := &store.Config{
		Logger: logger,
		DSN:    v.GetURL("database.dsn"),
	}
	svcCfg := &user.ServiceConfig{
		Logger:    logger,
		APIPrefix: v.GetURIPrefix("api.prefix"),
		AuthConfig: auth.Config{
			Secret:       v.GetString("auth.jwt.secret"),
			Expiration:   v.GetDuration("auth.jwt.expiration"),
			Domain:       v.GetString("domain"),
			SecureCookie: v.GetBool("https.enable"),
		},
	}
	srvCfg := &server.Config{
		Logger:            logger,
		ListenAddress:     v.GetString("server.http.address"),
		ReadTimeout:       v.GetDuration("server.http.timeouts.read"),
		ReadHeaderTimeout: v.GetDuration("server.http.timeouts.readHeader"),
		WriteTimeout:      v.GetDuration("server.http.timeouts.write"),
		IdleTimeout:       v.GetDuration("server.http.timeouts.idle"),
	}
	if v.HasErrors() {
		for param, errP := range v.Errors() {
			logger.Error("configuration error", zap.String("param", param), zap.Error(errP))
		}
		return nil, nil, false
	}

	userStorage, errStore := store.New(ctx, storeCfg)
	if errStore != nil {
		logger.Error("could not configure api storage", zap.Error(errStore))
		return nil, nil, false
	}
	svcCfg.Store = userStorage
	svc, errSvc := user.NewService(svcCfg)
	if errSvc != nil {
		logger.Error("could not configure api service", zap.Error(errSvc))
		return nil, nil, false
	}
	srv, errSrv := server.NewServer(srvCfg)
	if errSrv != nil {
		logger.Error("could not configure server", zap.Error(errSrv))
		return nil, nil, false
	}
	srv.SetHandler(svc)
	return []app.Runner{srv}, []app.Closer{userStorage}, true
}
