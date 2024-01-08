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
	defaultShutdownTimeout = 10 * time.Second
)

type Server struct {
	logger *zap.Logger
	srv    *http.Server
}

type Config struct {
	Logger            *zap.Logger
	ListenAddress     string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

func NewServer(cfg *Config) (*Server, error) {
	if cfg.Logger == nil {
		return nil, errors.New("nil logger")
	}

	return &Server{
		logger: cfg.Logger.With(zap.String("component", "server")),
		srv: &http.Server{
			Addr:              cfg.ListenAddress,
			ReadTimeout:       cfg.ReadTimeout,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		},
	}, nil
}

func (s *Server) SetHandler(h http.Handler) {
	s.srv.Handler = h
}

func (s *Server) Run(ctx context.Context, wg *sync.WaitGroup, errc chan<- error) {
	defer wg.Done()
	if s.srv.Handler == nil {
		errc <- errors.New("server handler is not set")
	}

	errSrv := make(chan error)
	go func() {
		errSrv <- s.srv.ListenAndServe()
	}()

	s.logger.Info("server started", zap.String("address", s.srv.Addr))

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
}
