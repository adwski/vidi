package grpc

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/adwski/vidi/internal/api/requestid"
	"github.com/adwski/vidi/internal/api/user/auth"
	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	defaultServerConnectionTimeout = 3 * time.Second

	defaultRPCTimeout = 5 * time.Second
)

// Server is videoapi GRPC server.
type Server struct {
	logger       *zap.Logger
	registerFunc func(s grpc.ServiceRegistrar)
	addr         string
	opts         []grpc.ServerOption
	reflection   bool
}

// Config is videoapi GRPC server config.
type Config struct {
	Logger     *zap.Logger
	Auth       *auth.Auth
	TLSConfig  *tls.Config
	ListenAddr string
	Reflection bool
}

// NewServer creates videoapi GRPC server.
func NewServer(cfg *Config, registerFunc func(s grpc.ServiceRegistrar)) (*Server, error) {
	if registerFunc == nil {
		return nil, errors.New("registerFunc cannot be nil")
	}
	if cfg.Auth == nil {
		return nil, errors.New("authenticator cannot be nil")
	}

	// Server params
	var opts []grpc.ServerOption
	opts = append(opts, grpc.ConnectionTimeout(defaultServerConnectionTimeout))
	if cfg.TLSConfig != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(cfg.TLSConfig)))
	}

	// assign interceptors
	opts = append(opts,
		// request ID
		grpc.ChainUnaryInterceptor(requestid.New(cfg.Logger, false).InterceptorFunc()),
		// logger
		grpc.ChainUnaryInterceptor(interceptorLogger(cfg.Logger)),
		// deadline
		grpc.ChainUnaryInterceptor(interceptorDeadline(defaultRPCTimeout)),
		// auth
		grpc.ChainUnaryInterceptor(grpcauth.UnaryServerInterceptor(cfg.Auth.GRPCAuthFunc)),
	)

	return &Server{
		logger:       cfg.Logger,
		addr:         cfg.ListenAddr,
		reflection:   cfg.Reflection,
		registerFunc: registerFunc,
		opts:         opts,
	}, nil
}

// Run starts grpc server and returns only when Listener stops (canceled).
// It should be started asynchronously and canceled via context.
// Error channel should be used to catch listen errors.
// If error is caught that means server is no longer running.
func (srv *Server) Run(ctx context.Context, wg *sync.WaitGroup, errc chan<- error) {
	defer wg.Done()

	listener, err := net.Listen("tcp", srv.addr)
	if err != nil {
		srv.logger.Error("cannot create listener", zap.Error(err))
		errc <- err
		return
	}

	// create grpc server
	s := grpc.NewServer(srv.opts...)
	srv.registerFunc(s)
	if srv.reflection {
		reflection.Register(s)
	}

	// start server
	srv.logger.Info("starting server", zap.String("address", listener.Addr().String()))
	errSrv := make(chan error)
	go func(errc chan<- error) {
		errc <- s.Serve(listener)
	}(errSrv)

	// wait for signals
	select {
	case <-ctx.Done():
	case err = <-errSrv:
		srv.logger.Error("listener error", zap.Error(err))
		errc <- err
	}
	s.GracefulStop()
	srv.logger.Info("server stopped")
}

// interceptorDeadline creates circuit breaking interceptor
// that limits request processing time.
//
// Time limits is guaranteed (compared with if we would've just passed deadline context).
func interceptorDeadline(deadline time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		newCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		var (
			resp interface{}
			err  error
			done = make(chan struct{})
		)
		go func() {
			resp, err = handler(newCtx, req)
			done <- struct{}{}
		}()
		select {
		case <-time.After(deadline):
			return nil, status.Error(codes.DeadlineExceeded, "rpc deadline")
		case <-done:
			return resp, err
		}
	}
}

// interceptorLogger creates zap-flavoured grpc logging interceptor.
// Taken from https://github.com/grpc-ecosystem/go-grpc-middleware/blob/main/interceptors/logging/examples/zap/example_test.go.
//
//nolint:lll // link
func interceptorLogger(l *zap.Logger) grpc.UnaryServerInterceptor {
	loggerFunc := func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2) //nolint:mnd // half

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).
			With(zap.String("component", "grpc")).
			With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			logger.Error("grpc logger called with unknown level", zap.Any("lvl", lvl))
		}
	}
	return logging.UnaryServerInterceptor(logging.LoggerFunc(loggerFunc))
}
