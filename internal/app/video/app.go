package video

import (
	"context"
	httpserver "github.com/adwski/vidi/internal/api/http/server"
	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/adwski/vidi/internal/api/video"
	"github.com/adwski/vidi/internal/api/video/http/server"
	"github.com/adwski/vidi/internal/api/video/store"
	"github.com/adwski/vidi/internal/app"
	"github.com/adwski/vidi/internal/session"
	sessionStore "github.com/adwski/vidi/internal/session/store"
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
		DSN:    v.GetString("database.dsn"),
	}
	svcCfg := &video.ServiceConfig{
		Logger:          logger,
		WatchURLPrefix:  v.GetURL("media.url.watch"),
		UploadURLPrefix: v.GetURL("media.url.upload"),
	}
	srvCfg := &server.Config{
		Logger:    logger,
		APIPrefix: v.GetURIPrefix("api.prefix"),
		AuthConfig: auth.Config{
			Secret:     v.GetString("auth.jwt.secret"),
			Expiration: v.GetDuration("auth.jwt.expiration"),
			Domain:     v.GetString("domain"),
			HTTPS:      v.GetBool("https.enable"),
		},
		HTTPConfig: &httpserver.Config{
			Logger:            logger,
			ListenAddress:     v.GetString("server.address"),
			ReadTimeout:       v.GetDuration("server.timeouts.read"),
			ReadHeaderTimeout: v.GetDuration("server.timeouts.readHeader"),
			WriteTimeout:      v.GetDuration("server.timeouts.write"),
			IdleTimeout:       v.GetDuration("server.timeouts.idle"),
		},
	}
	uploadSessionStoreCfg := &sessionStore.Config{
		Logger:   logger,
		Name:     session.KindUpload,
		RedisDSN: v.GetURL("redis.dsn"),
		TTL:      v.GetDuration("redis.ttl.upload"),
	}
	watchSessionStoreCfg := &sessionStore.Config{
		Logger:   logger,
		Name:     session.KindWatch,
		RedisDSN: v.GetURL("redis.dsn"),
		TTL:      v.GetDuration("redis.ttl.watch"),
	}
	if v.HasErrors() {
		for param, errP := range v.Errors() {
			logger.Error("configuration error", zap.String("param", param), zap.Error(errP))
		}
		return nil, nil, false
	}
	uploadSessStore, errSSU := sessionStore.NewStore(uploadSessionStoreCfg)
	if errSSU != nil {
		logger.Error("cannot configure upload session store", zap.Error(errSSU))
		return nil, nil, false
	}
	watchSessStore, errSSW := sessionStore.NewStore(watchSessionStoreCfg)
	if errSSW != nil {
		logger.Error("cannot configure watch session store", zap.Error(errSSW))
		return nil, nil, false
	}
	videoStorage, errStore := store.New(ctx, storeCfg)
	if errStore != nil {
		logger.Error("could not configure api storage", zap.Error(errStore))
		return nil, nil, false
	}
	svcCfg.UploadSessionStore = uploadSessStore
	svcCfg.WatchSessionStore = watchSessStore
	svcCfg.Store = videoStorage

	srv, errSrv := server.NewServer(srvCfg, video.NewService(svcCfg))
	if errSrv != nil {
		logger.Error("could not configure server", zap.Error(errSrv))
		return nil, nil, false
	}
	return []app.Runner{srv}, []app.Closer{videoStorage, uploadSessStore, watchSessStore}, true
}
