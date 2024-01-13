package processor

import (
	"context"

	"github.com/adwski/vidi/internal/app"
	"github.com/adwski/vidi/internal/media/processor"
	"github.com/adwski/vidi/internal/media/store/s3"
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
	processorCfg := &processor.Config{
		Logger:           logger,
		Store:            nil,
		VideoAPIEndpoint: v.GetURL("videoapi.endpoint"),
		VideoAPIToken:    v.GetString("videoapi.token"),
		InputPathPrefix:  v.GetURIPrefix("s3.prefix.upload"),
		OutputPathPrefix: v.GetURIPrefix("s3.prefix.watch"),
		SegmentDuration:  v.GetDuration("processor.segment_duration"),
		VideoCheckPeriod: v.GetDuration("processor.video_check_period"),
	}
	storageCfg := &s3.StoreConfig{
		Logger:    logger,
		Endpoint:  v.GetString("s3.endpoint"),
		AccessKey: v.GetString("s3.access_key"),
		SecretKey: v.GetString("s3.secret_key"),
		Bucket:    v.GetString("s3.bucket"),
		SSL:       v.GetBool("s3.ssl"),
	}
	if v.HasErrors() {
		for param, errP := range v.Errors() {
			logger.Error("configuration error", zap.String("param", param), zap.Error(errP))
		}
		return nil, nil, false
	}
	store, err := s3.NewStore(storageCfg)
	if err != nil {
		logger.Error("cannot create s3 storage", zap.Error(err))
		return nil, nil, false
	}
	processorCfg.Store = store
	return processor.New(processorCfg), nil, true
}
