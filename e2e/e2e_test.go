//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"github.com/adwski/vidi/internal/app/processor"
	"github.com/adwski/vidi/internal/app/streamer"
	"github.com/adwski/vidi/internal/app/uploader"
	"github.com/adwski/vidi/internal/app/video"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/adwski/vidi/internal/app/user"
)

func TestMain(m *testing.M) {
	var (
		wg          = &sync.WaitGroup{}
		ctx, cancel = context.WithCancel(context.Background())
	)

	wg.Add(5)
	go func() {
		user.NewApp().RunWithContextAndConfig(ctx, "userapi.yaml")
		wg.Done()
	}()
	go func() {
		video.NewApp().RunWithContextAndConfig(ctx, "videoapi.yaml")
		wg.Done()
	}()
	go func() {
		uploader.NewApp().RunWithContextAndConfig(ctx, "uploader.yaml")
		wg.Done()
	}()
	go func() {
		processor.NewApp().RunWithContextAndConfig(ctx, "processor.yaml")
		wg.Done()
	}()
	go func() {
		streamer.NewApp().RunWithContextAndConfig(ctx, "streamer.yaml")
		wg.Done()
	}()

	time.Sleep(3 * time.Second)

	code := m.Run()
	cancel()
	wg.Wait()
	defer func() {
		os.Exit(code)
	}()
}
