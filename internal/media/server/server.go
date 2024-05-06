package server

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

const (
	defaultShutdownTimeout = 15 * time.Second

	defaultMaxBodySize = 1 * 1024 * 1024
)

// Server is a runnable fasthttp server intended to be used with media apps.
type Server struct {
	logger *zap.Logger
	srv    *fasthttp.Server
	addr   string
}

type Config struct {
	Logger        *zap.Logger
	Handler       fasthttp.RequestHandler
	MaxBodySize   int
	ListenAddress string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
}

func New(cfg *Config) *Server {
	if cfg.MaxBodySize < defaultMaxBodySize {
		cfg.MaxBodySize = defaultMaxBodySize
	}
	return &Server{
		logger: cfg.Logger.With(zap.String("component", "server")),
		addr:   cfg.ListenAddress,
		srv: &fasthttp.Server{
			Handler:            cfg.Handler,
			ReadTimeout:        cfg.ReadTimeout,
			WriteTimeout:       cfg.WriteTimeout,
			IdleTimeout:        cfg.IdleTimeout,
			MaxRequestBodySize: cfg.MaxBodySize,
			CloseOnShutdown:    true,
			//StreamRequestBody:  true,

			// Prevent handling of potentially large requests
			// by forbidding all 100s, since we're not using them now.
			// If client sends 100-Continue, server will respond with 417
			ContinueHandler: func(_ *fasthttp.RequestHeader) bool { return false },
		}}
}

func (s *Server) Run(ctx context.Context, wg *sync.WaitGroup, errc chan<- error) {
	defer wg.Done()
	if s.srv.Handler == nil {
		errc <- errors.New("server handler is not set")
		return
	}

	errSrv := make(chan error)
	go func() {
		errSrv <- s.srv.ListenAndServe(s.addr)
	}()

	s.logger.Info("server started", zap.String("address", s.addr))

	select {
	case <-ctx.Done():
		ctxSh, cancelSh := context.WithDeadline(context.Background(), time.Now().Add(defaultShutdownTimeout))
		defer cancelSh()
		// Read and write timeouts should be set to something sane
		// for graceful shutdown to return in reasonable time
		if err := s.srv.ShutdownWithContext(ctxSh); err != nil {
			s.logger.Error("error during server shutdown", zap.Error(err))
		}

	case err := <-errSrv:
		if !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("server error", zap.Error(err))
			errc <- err
		}
	}
}
