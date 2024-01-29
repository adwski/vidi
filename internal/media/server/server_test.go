package server

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func TestServer_RunNoHandler(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	s := New(&Config{
		Logger:        logger,
		ListenAddress: ":8881",
		ReadTimeout:   time.Second,
		WriteTimeout:  time.Second,
		IdleTimeout:   time.Second,
	})

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

	s := New(&Config{
		Handler:       stub,
		Logger:        logger,
		ListenAddress: ":8881",
		ReadTimeout:   time.Second,
		WriteTimeout:  time.Second,
		IdleTimeout:   time.Second,
	})

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

	s := New(&Config{
		Handler:       stub,
		Logger:        logger,
		ListenAddress: ":888888",
		ReadTimeout:   time.Second,
		WriteTimeout:  time.Second,
		IdleTimeout:   time.Second,
	})

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

var stub = func(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
}
