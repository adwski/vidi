package uploader

import (
	"context"

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

func (a *App) configure(_ context.Context) (app.Runner, app.Closer, bool) {
	var (
		logger = a.Logger()
		v      = a.Viper()
	)

	uploaderCfg := uploader.Config{
		Logger:        logger,
		RedisDSN:      v.GetString("redis.dsn"),
		URIPathPrefix: v.GetString("api.prefix"),
		VideoAPIURL:   v.GetString("videoapi.endpoint"),
		VideoAPIToken: v.GetString("videoapi.token"),
		S3PathPrefix:  v.GetString("s3.prefix.upload"),
		S3Endpoint:    v.GetString("s3.endpoint"),
		S3AccessKey:   v.GetString("s3.access_key"),
		S3SecretKey:   v.GetString("s3.secret_key"),
		S3Bucket:      v.GetString("s3.bucket"),
		S3SSL:         v.GetBool("s3.ssl"),
	}
	srvCfg := &server.Config{
		Logger:        logger,
		ListenAddress: v.GetString("server.address"),
		ReadTimeout:   v.GetDuration("server.timeouts.read"),
		WriteTimeout:  v.GetDuration("server.timeouts.write"),
		IdleTimeout:   v.GetDuration("server.timeouts.idle"),
	}
	if v.HasErrors() {
		for param, errP := range v.Errors() {
			logger.Error("configuration error", zap.String("param", param), zap.Error(errP))
		}
		return nil, nil, false
	}
	uploaderSvc, errUp := uploader.New(&uploaderCfg)
	if errUp != nil {
		logger.Error("cannot create uploader service", zap.Error(errUp))
		return nil, nil, false
	}
	srvCfg.Handler = uploaderSvc.Handler()
	return server.New(srvCfg), nil, true
}
