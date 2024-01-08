package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/adwski/vidi/internal/api/server"
	"github.com/adwski/vidi/internal/logging"
	"github.com/adwski/vidi/pkg/config"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	envPrefix = "VIDI"

	defaultReadHeaderTimeout = time.Second
	defaultReadTimeout       = 5 * time.Second
	defaultWriteTimeout      = 5 * time.Second
	defaultIdleTimeout       = 10 * time.Second

	defaultJWTExpiration = 12 * time.Hour
)

type Closer interface {
	Close()
}

type App struct {
	defaultLogger *zap.Logger
	logger        *zap.Logger
	viper         *config.ViperEC
	srv           *server.Server
	svcInitFunc   func(ctx context.Context) (http.Handler, Closer, bool)
}

func New(initFunc func(ctx context.Context) (http.Handler, Closer, bool)) *App {
	return &App{
		defaultLogger: logging.GetZapLoggerDefaultLevel(),
		viper:         config.NewViperEC(),
		svcInitFunc:   initFunc,
	}
}

func (app *App) Logger() *zap.Logger {
	return app.logger
}

func (app *App) Viper() *config.ViperEC {
	return app.viper
}

func (app *App) Run() int {
	// --------------------------------------------------
	// configure
	// --------------------------------------------------
	code := app.configure()
	if code != 0 {
		return code
	}

	// --------------------------------------------------
	// start app
	// --------------------------------------------------
	return app.run()
}

func (app *App) run() int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Configure user api service
	handler, closer, ok := app.svcInitFunc(ctx)
	if !ok {
		return 1
	}
	app.srv.SetHandler(handler)

	var (
		wg   = &sync.WaitGroup{}
		errc = make(chan error)
	)
	wg.Add(1)
	go app.srv.Run(ctx, wg, errc)

	select {
	case <-ctx.Done():
		app.logger.Warn("shutting down")
	case <-errc:
		cancel()
	}
	wg.Wait()
	closer.Close()
	return 0
}

func (app *App) configure() int {
	// Set defaults
	app.setConfigDefaults()
	// Try to read the config ignoring any errors
	err := app.readConfig()
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

	// Configure server
	srvCfg := &server.Config{
		Logger:            app.logger,
		ListenAddress:     app.viper.GetString("server.address"),
		ReadTimeout:       app.viper.GetDuration("server.timeouts.read"),
		ReadHeaderTimeout: app.viper.GetDuration("server.timeouts.readHeader"),
		WriteTimeout:      app.viper.GetDuration("server.timeouts.write"),
		IdleTimeout:       app.viper.GetDuration("server.timeouts.idle"),
	}
	if app.viper.HasErrors() {
		for param, errP := range app.viper.Errors() {
			app.logger.Error("server param error", zap.String("param", param), zap.Error(errP))
		}
		return 1
	}
	app.srv, err = server.NewServer(srvCfg)
	if err != nil {
		app.logger.Error("could not configure server", zap.Error(err))
		return 1
	}
	return 0
}

func (app *App) readConfig() error {
	app.viper.SetConfigName("config")
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
	v.SetDefault("server.address", ":8080")
	v.SetDefault("server.timeouts.readHeader", defaultReadHeaderTimeout)
	v.SetDefault("server.timeouts.read", defaultReadTimeout)
	v.SetDefault("server.timeouts.write", defaultWriteTimeout)
	v.SetDefault("server.timeouts.idle", defaultIdleTimeout)
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
	v.SetDefault("database.salt", "changeMe")

	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}
