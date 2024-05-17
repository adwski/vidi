package streamer

import (
	"context"

	"github.com/adwski/vidi/internal/app"
	"github.com/adwski/vidi/internal/media/server"
	"github.com/adwski/vidi/internal/media/streamer"
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

func (a *App) configure(_ context.Context) ([]app.Runner, []app.Closer, bool) {
	var (
		logger = a.Logger()
		v      = a.Viper()
	)

	var corsConfig *streamer.CORSConfig
	if v.GetBoolNoError("cors.enable") {
		corsConfig = &streamer.CORSConfig{AllowOrigin: v.GetString("cors.allow_origin")}
	}

	streamerCfg := streamer.Config{
		Logger:        logger,
		CORSConfig:    corsConfig,
		URIPathPrefix: v.GetURIPrefix("api.prefix"),
		S3PathPrefix:  v.GetURIPrefix("s3.prefix.watch"),
		S3Endpoint:    v.GetString("s3.endpoint"),
		S3AccessKey:   v.GetString("s3.access_key"),
		S3SecretKey:   v.GetString("s3.secret_key"),
		S3Bucket:      v.GetString("s3.bucket"),
		S3SSL:         v.GetBool("s3.ssl"),
	}
	srvCfg := &server.Config{
		Logger:        logger,
		ListenAddress: v.GetString("server.http.address"),
		ReadTimeout:   v.GetDuration("server.http.timeouts.read"),
		WriteTimeout:  v.GetDuration("server.http.timeouts.write"),
		IdleTimeout:   v.GetDuration("server.http.timeouts.idle"),
	}
	sessionStoreCfg := &sessionStore.Config{
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
	sessStore, errSS := sessionStore.NewStore(sessionStoreCfg)
	if errSS != nil {
		logger.Error("cannot configure session store", zap.Error(errSS))
		return nil, nil, false
	}
	streamerCfg.SessionStore = sessStore
	streamerSvc, errUp := streamer.New(&streamerCfg)
	if errUp != nil {
		logger.Error("cannot create uploader service", zap.Error(errUp))
		return nil, nil, false
	}
	srvCfg.Handler = streamerSvc.Handler()
	return []app.Runner{server.New(srvCfg)}, []app.Closer{sessStore}, true
}
