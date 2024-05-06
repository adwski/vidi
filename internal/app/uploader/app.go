package uploader

import (
	"context"

	"github.com/adwski/vidi/internal/event/notificator"
	"github.com/adwski/vidi/internal/session"
	sessionStore "github.com/adwski/vidi/internal/session/store"

	"github.com/adwski/vidi/internal/app"
	"github.com/adwski/vidi/internal/media/server"
	"github.com/adwski/vidi/internal/media/uploader"
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

	uploaderCfg := uploader.Config{
		Logger:        logger,
		URIPathPrefix: v.GetURIPrefix("api.prefix"),
		S3PathPrefix:  v.GetURIPrefix("s3.prefix.upload"),
		S3Endpoint:    v.GetString("s3.endpoint"),
		S3AccessKey:   v.GetString("s3.access_key"),
		S3SecretKey:   v.GetString("s3.secret_key"),
		S3Bucket:      v.GetString("s3.bucket"),
		S3SSL:         v.GetBool("s3.ssl"),
	}
	notificatorCfg := &notificator.Config{
		Logger:        logger,
		VideoAPIURL:   v.GetURL("videoapi.endpoint"),
		VideoAPIToken: v.GetString("videoapi.token"),
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
		Name:     session.KindUpload,
		RedisDSN: v.GetURL("redis.dsn"),
		TTL:      v.GetDuration("redis.ttl.upload"),
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
	uploaderCfg.SessionStorage = sessStore
	uploaderCfg.Notificator = notificator.New(notificatorCfg)
	uploaderSvc, errUp := uploader.New(&uploaderCfg)
	if errUp != nil {
		logger.Error("cannot create uploader service", zap.Error(errUp))
		return nil, nil, false
	}
	srvCfg.Handler = uploaderSvc.Handler()
	return []app.Runner{server.New(srvCfg), uploaderCfg.Notificator}, []app.Closer{sessStore}, true
}
