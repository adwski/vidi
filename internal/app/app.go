package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/adwski/vidi/internal/logging"
	"github.com/adwski/vidi/pkg/config"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	envPrefix = "VIDI"

	defaultConfigName = "config"

	defaultReadHeaderTimeout = time.Second
	defaultReadTimeout       = 5 * time.Second
	defaultWriteTimeout      = 5 * time.Second
	defaultIdleTimeout       = 10 * time.Second

	defaultJWTExpiration = 12 * time.Hour

	defaultSegmentDuration    = 3 * time.Second
	defaultVideoCheckInterval = 5 * time.Second

	defaultUploadSessionTTL = 300 * time.Second
	defaultWatchSessionTTL  = 600 * time.Second

	defaultMaxVideos = 100
	defaultMaxSize   = 10 * 1 << 30
)

type Runner interface {
	Run(ctx context.Context, wg *sync.WaitGroup, errc chan<- error)
}

type Closer interface {
	Close()
}

type Initializer func(context.Context) ([]Runner, []Closer, bool)

// App is a common skeleton for daemon application.
type App struct {
	defaultLogger *zap.Logger
	logger        *zap.Logger
	viper         *config.ViperEC
	initializer   Initializer
}

func New(initializer Initializer) *App {
	return &App{
		defaultLogger: logging.GetZapLoggerDefaultLevel(),
		viper:         config.NewViperEC(),
		initializer:   initializer,
	}
}

func (app *App) Logger() *zap.Logger {
	return app.logger
}

func (app *App) Viper() *config.ViperEC {
	return app.viper
}

func (app *App) Run() int {
	return app.RunWithContextAndConfig(context.Background(), defaultConfigName)
}

func (app *App) RunWithContextAndConfig(ctx context.Context, configFileName string) int {
	// --------------------------------------------------
	// configure
	// --------------------------------------------------
	code := app.configure(configFileName)
	if code != 0 {
		return code
	}

	// --------------------------------------------------
	// start app
	// --------------------------------------------------
	return app.run(ctx)
}

func (app *App) run(ctx context.Context) int {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	runners, closers, ok := app.initializer(ctx)
	if !ok {
		return 1
	}

	var (
		wg   = &sync.WaitGroup{}
		errc = make(chan error)
	)
	for _, r := range runners {
		wg.Add(1)
		go r.Run(ctx, wg, errc)
	}

	select {
	case <-ctx.Done():
		app.logger.Warn("shutting down")
	case <-errc:
		cancel()
	}
	wg.Wait()
	for _, c := range closers {
		c.Close()
	}
	return 0
}

func (app *App) configure(configName string) int {
	// Set defaults
	app.setConfigDefaults()
	// Try to read the config ignoring any errors
	err := app.readConfig(configName)
	if err != nil {
		app.defaultLogger.Error("config error", zap.Error(err))
		return 1
	}
	// Get logger with specified level
	app.logger, err = logging.GetZapLoggerWithLevel(app.viper.GetString("log.level"))
	if err != nil {
		app.defaultLogger.Error("could not parse log level", zap.Error(err))
		return 1
	}
	return 0
}

func (app *App) SetLogger(logger *zap.Logger) {
	app.logger = logger
}

func (app *App) readConfig(name string) error {
	app.viper.SetConfigName(name)
	app.viper.SetConfigType("yaml")
	app.viper.AddConfigPath(".")
	if err := app.viper.ReadInConfig(); err != nil {
		var vNotFound viper.ConfigFileNotFoundError
		if ok := errors.As(err, &vNotFound); !ok {
			return fmt.Errorf("error while reading config file: %w", err)
		}
	}
	return nil
}

func (app *App) setConfigDefaults() {
	// Set default config params
	v := app.viper
	// Logging
	v.SetDefault("log.level", "debug")
	// Server
	v.SetDefault("server.http.address", ":8080")
	v.SetDefault("server.http.timeouts.readHeader", defaultReadHeaderTimeout)
	v.SetDefault("server.http.timeouts.read", defaultReadTimeout)
	v.SetDefault("server.http.timeouts.write", defaultWriteTimeout)
	v.SetDefault("server.http.timeouts.idle", defaultIdleTimeout)
	// GRPC
	v.SetDefault("server.grpc.address", ":8181")
	v.SetDefault("server.grpc.svc_address", ":8282")
	v.SetDefault("server.grpc.reflection", false)
	// TLS
	v.SetDefault("server.tls.enable", false)
	v.SetDefault("server.tls.cert", "")
	v.SetDefault("server.tls.key", "")
	// Redis
	v.SetDefault("redis.dsn", "redis://localhost:6379/0")
	v.SetDefault("redis.ttl.upload", defaultUploadSessionTTL)
	v.SetDefault("redis.ttl.watch", defaultWatchSessionTTL)
	// S3
	v.SetDefault("s3.prefix.upload", "/")
	v.SetDefault("s3.prefix.watch", "/")
	v.SetDefault("s3.endpoint", "minio:9000")
	v.SetDefault("s3.bucket", "vidi")
	v.SetDefault("s3.access_key", "access-key")
	v.SetDefault("s3.secret_key", "secret-key")
	v.SetDefault("s3.ssl", false)
	// Video API Client
	v.SetDefault("videoapi.token", "changeMe")
	v.SetDefault("videoapi.endpoint", "http://videoapi:8080/api/video")
	// Common
	v.SetDefault("domain", "localhost")
	v.SetDefault("https.enable", false)
	// Auth
	v.SetDefault("auth.jwt.secret", "changeMe")
	v.SetDefault("auth.jwt.expiration", defaultJWTExpiration)
	// API
	v.SetDefault("api.prefix", "/api")
	// DB
	v.SetDefault("database.dsn", "postgres://postgres:postgres@localhost:5432/postgres")
	// Processor
	v.SetDefault("processor.segment_duration", defaultSegmentDuration)
	v.SetDefault("processor.video_check_period", defaultVideoCheckInterval)
	// Media
	v.SetDefault("media.user_quota.max_videos", defaultMaxVideos)
	v.SetDefault("media.user_quota.max_size", defaultMaxSize)

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}
