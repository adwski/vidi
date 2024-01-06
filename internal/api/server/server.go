package server

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	defaultReadHeaderTimeout = time.Second
	defaultReadTimeout       = 5 * time.Second
	defaultWriteTimeout      = 5 * time.Second
	defaultIdleTimeout       = 10 * time.Second
	defaultShutdownTimeout   = 10 * time.Second
)

type Server struct {
	logger *zap.Logger
	srv    *http.Server
}

type Config struct {
	Logger        *zap.Logger
	Handler       http.Handler
	ListenAddress string
}

func NewServer(cfg *Config) (*Server, error) {
	if cfg.Logger == nil {
		return nil, errors.New("nil logger")
	}

	return &Server{
		logger: cfg.Logger,
		srv: &http.Server{
			Addr:              cfg.ListenAddress,
			Handler:           cfg.Handler,
			ReadTimeout:       defaultReadTimeout,
			ReadHeaderTimeout: defaultReadHeaderTimeout,
			WriteTimeout:      defaultWriteTimeout,
			IdleTimeout:       defaultIdleTimeout,
		},
	}, nil
}

func (s *Server) Run(ctx context.Context, wg *sync.WaitGroup, errc chan<- error) {
	errSrv := make(chan error)
	go func() {
		errSrv <- s.srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		ctxSh, cancelSh := context.WithDeadline(context.Background(), time.Now().Add(defaultShutdownTimeout))
		defer cancelSh()
		if err := s.srv.Shutdown(ctxSh); err != nil {
			s.logger.Error("error during server shutdown", zap.Error(err))
		}

	case err := <-errSrv:
		if !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("server error", zap.Error(err))
			errc <- err
		}
	}
	wg.Done()
}
