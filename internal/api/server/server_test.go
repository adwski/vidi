package server

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewServerNilLogger(t *testing.T) {
	_, errS := NewServer(&Config{
		ListenAddress:     ":8888",
		ReadTimeout:       time.Second,
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      time.Second,
		IdleTimeout:       time.Second,
	})
	assert.ErrorContains(t, errS, "nil logger")
}

func TestServer_RunNoHandler(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s, errS := NewServer(&Config{
		Logger:            logger,
		ListenAddress:     ":8881",
		ReadTimeout:       time.Second,
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      time.Second,
		IdleTimeout:       time.Second,
	})
	require.NoError(t, errS)

	var (
		ctx, cancel = context.WithCancel(context.Background())
		wg          = &sync.WaitGroup{}
		errc        = make(chan error)
	)

	wg.Add(1)
	go s.Run(ctx, wg, errc)

	select {
	case err := <-errc:
		require.ErrorContains(t, err, "server handler is not set")
	case <-time.After(time.Second):
		assert.Fail(t, "no error was returned")
	}
	cancel()
	wg.Wait()
}

func TestServer_RunCancel(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s, errS := NewServer(&Config{
		Logger:            logger,
		ListenAddress:     ":8888",
		ReadTimeout:       time.Second,
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      time.Second,
		IdleTimeout:       time.Second,
	})
	require.NoError(t, errS)

	s.SetHandler(stub{})

	var (
		ctx, cancel = context.WithCancel(context.Background())
		wg          = &sync.WaitGroup{}
		errc        = make(chan error)
		done        = make(chan struct{})
	)

	wg.Add(1)
	go s.Run(ctx, wg, errc)

	go func() {
		cancel()
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		assert.Fail(t, "did not shutdown in time")
	}
}

func TestServer_RunError(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s, errS := NewServer(&Config{
		Logger:            logger,
		ListenAddress:     ":888888",
		ReadTimeout:       time.Second,
		ReadHeaderTimeout: time.Second,
		WriteTimeout:      time.Second,
		IdleTimeout:       time.Second,
	})
	require.NoError(t, errS)

	s.SetHandler(stub{})

	var (
		ctx, cancel = context.WithCancel(context.Background())
		wg          = &sync.WaitGroup{}
		errc        = make(chan error)
	)
	defer cancel()

	wg.Add(1)
	go s.Run(ctx, wg, errc)

	select {
	case err := <-errc:
		assert.Error(t, err)
	case <-time.After(time.Second):
		assert.Fail(t, "did not return error")
	}
	wg.Wait()
}

type stub struct{}

func (s stub) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
