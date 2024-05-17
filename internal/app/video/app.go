package video

import (
	"context"
	"crypto/tls"

	httpserver "github.com/adwski/vidi/internal/api/http/server"
	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/adwski/vidi/internal/api/video"
	"github.com/adwski/vidi/internal/api/video/grpc"
	"github.com/adwski/vidi/internal/api/video/grpc/serviceside"
	"github.com/adwski/vidi/internal/api/video/grpc/userside"
	"github.com/adwski/vidi/internal/api/video/http/server"
	"github.com/adwski/vidi/internal/api/video/store"
	"github.com/adwski/vidi/internal/app"
	"github.com/adwski/vidi/internal/session"
	sessionStore "github.com/adwski/vidi/internal/session/store"
	"go.uber.org/zap"
)

const (
	minTLSVersion = tls.VersionTLS13
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
	// --------------------------------------
	// Gather config params
	// --------------------------------------
	storeCfg := &store.Config{
		Logger: logger,
		DSN:    v.GetString("database.dsn"),
	}
	svcCfg := &video.ServiceConfig{
		Logger:          logger,
		WatchURLPrefix:  v.GetURL("media.url.watch"),
		UploadURLPrefix: v.GetURL("media.url.upload"),
		Quotas: video.Quotas{
			VideosPerUser: v.GetUint("media.user_quota.max_videos"),
			MaxTotalSize:  v.GetUint64("media.user_quota.max_size"),
		},
	}
	authCfg := auth.Config{
		Secret:       v.GetString("auth.jwt.secret"),
		Expiration:   v.GetDuration("auth.jwt.expiration"),
		Domain:       v.GetString("domain"),
		SecureCookie: v.GetBool("https.enable"),
	}
	srvCfg := &server.Config{
		Logger:    logger,
		APIPrefix: v.GetURIPrefix("api.prefix"),
		HTTPConfig: &httpserver.Config{
			Logger:            logger,
			ListenAddress:     v.GetString("server.http.address"),
			ReadTimeout:       v.GetDuration("server.http.timeouts.read"),
			ReadHeaderTimeout: v.GetDuration("server.http.timeouts.readHeader"),
			WriteTimeout:      v.GetDuration("server.http.timeouts.write"),
			IdleTimeout:       v.GetDuration("server.http.timeouts.idle"),
		},
	}
	gUserSrvCfg := &grpc.Config{
		Logger:     logger,
		ListenAddr: v.GetString("server.grpc.address"),
		Reflection: v.GetBool("server.grpc.reflection"),
	}
	gServiceSrvCfg := &grpc.Config{
		Logger:     logger,
		ListenAddr: v.GetString("server.grpc.svc_address"),
		Reflection: v.GetBool("server.grpc.reflection"),
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
	var (
		tlsKeyPath, tlsCertPath string
		grpcTLSEnableUsr        = v.GetBool("server.grpc.tls_userside_enable")
		grpcTLSEnableSvc        = v.GetBool("server.grpc.tls_serviceside_enable")
	)
	if grpcTLSEnableSvc || grpcTLSEnableUsr {
		tlsCertPath = v.GetString("server.tls.cert")
		tlsKeyPath = v.GetString("server.tls.key")
	}
	if v.HasErrors() {
		for param, errP := range v.Errors() {
			logger.Error("configuration error", zap.String("param", param), zap.Error(errP))
		}
		return nil, nil, false
	}

	// --------------------------------------
	// Spawn application entities
	// --------------------------------------
	// tls config
	if grpcTLSEnableSvc || grpcTLSEnableUsr {
		cert, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
		if err != nil {
			logger.Error("cannot create tls config", zap.Error(err))
			return nil, nil, false
		}
		if grpcTLSEnableUsr {
			gUserSrvCfg.TLSConfig = &tls.Config{
				MinVersion:   minTLSVersion,
				Certificates: []tls.Certificate{cert},
			}
		}
		if grpcTLSEnableSvc {
			gServiceSrvCfg.TLSConfig = &tls.Config{
				MinVersion:   minTLSVersion,
				Certificates: []tls.Certificate{cert},
			}
		}
	}

	// authenticator
	authenticator, errAuth := auth.NewAuth(&authCfg)
	if errAuth != nil {
		logger.Error("could not configure authenticator", zap.Error(errAuth))
		return nil, nil, false
	}
	gUserSrvCfg.Auth = authenticator
	gServiceSrvCfg.Auth = authenticator
	srvCfg.Auth = authenticator

	// upload session storage
	uploadSessStore, errSSU := sessionStore.NewStore(uploadSessionStoreCfg)
	if errSSU != nil {
		logger.Error("cannot configure upload session store", zap.Error(errSSU))
		return nil, nil, false
	}
	svcCfg.UploadSessionStore = uploadSessStore

	// watch session storage
	watchSessStore, errSSW := sessionStore.NewStore(watchSessionStoreCfg)
	if errSSW != nil {
		logger.Error("cannot configure watch session store", zap.Error(errSSW))
		return nil, nil, false
	}
	svcCfg.WatchSessionStore = watchSessStore

	// video storage
	videoStorage, errStore := store.New(ctx, storeCfg)
	if errStore != nil {
		logger.Error("could not configure api storage", zap.Error(errStore))
		return nil, nil, false
	}
	svcCfg.Store = videoStorage

	// video service
	svc := video.NewService(svcCfg)

	// video http server (userside)
	srv, errSrv := server.NewServer(srvCfg, svc)
	if errSrv != nil {
		logger.Error("could not create http server", zap.Error(errSrv))
		return nil, nil, false
	}

	// video grpc userside server
	gUserSrv, errGSrv := userside.NewServer(gUserSrvCfg, svc)
	if errGSrv != nil {
		logger.Error("could not create grpc userside server", zap.Error(errGSrv))
		return nil, nil, false
	}

	// video grpc serviceside server
	gServiceSrv, errGService := serviceside.NewServer(gServiceSrvCfg, svc)
	if errGService != nil {
		logger.Error("could not create grpc serviceside server", zap.Error(errGSrv))
		return nil, nil, false
	}

	// --------------------------------------
	// Return initialized entities
	// --------------------------------------
	return []app.Runner{srv, gUserSrv, gServiceSrv}, []app.Closer{videoStorage, uploadSessStore, watchSessStore}, true
}
