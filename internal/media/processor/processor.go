// Package processor contains mp4 file processor which can only do segmentation for now.
package processor

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/adwski/vidi/internal/api/video/client"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/event"
	"github.com/adwski/vidi/internal/event/notificator"
	"go.uber.org/zap"
)

const (
	defaultMediaStoreArtifactName = "artifact.mp4"
)

type MediaStore interface {
	Put(ctx context.Context, name string, r io.Reader, size int64) error
	Get(ctx context.Context, name string) (io.ReadCloser, int64, error)
}

type Processor struct {
	logger           *zap.Logger
	videoAPI         *client.Client
	notificator      *notificator.Notificator
	st               MediaStore
	inputPathPrefix  string
	outputPathPrefix string
	segmentDuration  time.Duration
	videoCheckPeriod time.Duration
}

type Config struct {
	Logger           *zap.Logger
	Store            MediaStore
	VideoAPIEndpoint string
	VideoAPIToken    string
	InputPathPrefix  string
	OutputPathPrefix string
	SegmentDuration  time.Duration
	VideoCheckPeriod time.Duration
}

func New(cfg *Config) *Processor {
	return &Processor{
		logger:           cfg.Logger.With(zap.String("component", "processor")),
		st:               cfg.Store,
		segmentDuration:  cfg.SegmentDuration,
		videoCheckPeriod: cfg.VideoCheckPeriod,
		inputPathPrefix:  fmt.Sprintf("%s/", strings.TrimSuffix(cfg.InputPathPrefix, "/")),
		outputPathPrefix: fmt.Sprintf("%s/", strings.TrimSuffix(cfg.OutputPathPrefix, "/")),
		notificator: notificator.New(&notificator.Config{
			Logger:        cfg.Logger,
			VideoAPIURL:   cfg.VideoAPIEndpoint,
			VideoAPIToken: cfg.VideoAPIToken,
		}),
		videoAPI: client.New(&client.Config{
			Logger:   cfg.Logger,
			Endpoint: cfg.VideoAPIEndpoint,
			Token:    cfg.VideoAPIToken,
		}),
	}
}

func (p *Processor) Run(ctx context.Context, wg *sync.WaitGroup, _ chan<- error) {
	defer wg.Done()
Loop:
	for {
		select {
		case <-ctx.Done():
			p.logger.Info("shutting down")
			break Loop
		case <-time.After(p.videoCheckPeriod):
			p.checkAndProcessVideos(ctx)
		}
	}
	p.logger.Info("stopped")
}

func (p *Processor) checkAndProcessVideos(ctx context.Context) {
	p.logger.Debug("checking videos")

	videos, err := p.videoAPI.GetUploadedVideos(ctx)
	if err != nil {
		p.logger.Error("cannot get videos from video API", zap.Error(err))
		return
	}
	if len(videos) == 0 {
		p.logger.Debug("no videos for processing")
		return
	}
	for _, video := range videos {
		p.notificator.Send(&event.Event{
			Video: model.Video{
				ID:     video.ID,
				Status: model.VideoStatusProcessing,
			},
			Kind: event.KindUpdateStatus,
		})
		if err = p.processVideo(ctx, video); err != nil {
			// TODO In the future we should distinguish between errors caused by video content
			//  and any other error. For example i/o errors are related to other failures
			//  and in such cases video processing could be retried later. (So we need retry mechanism).
			p.notificator.Send(&event.Event{
				Video: model.Video{
					ID:     video.ID,
					Status: model.VideoStatusError,
				},
				Kind: event.KindUpdateStatus,
			})
			p.logger.Error("error while processing video",
				zap.String("id", video.ID),
				zap.String("location", video.Location),
				zap.Error(err))
		}
		p.logger.Debug("video processed successfully",
			zap.String("id", video.ID),
			zap.String("location", video.Location))
		p.notificator.Send(&event.Event{
			Video: model.Video{
				ID:     video.ID,
				Status: model.VideoStatusReady,
			},
			Kind: event.KindUpdateStatus,
		})
	}
	p.logger.Debug("processing done")
}

func (p *Processor) processVideo(ctx context.Context, video *model.Video) error {
	fullInputPath := fmt.Sprintf("%s/%s/%s", p.inputPathPrefix, video.Location, defaultMediaStoreArtifactName)
	rc, _, err := p.st.Get(ctx, fullInputPath)
	if err != nil {
		return fmt.Errorf("cannot get input file: %w", err)
	}
	defer func() {
		if errC := rc.Close(); errC != nil {
			p.logger.Error("error closing storage reader", zap.Error(errC))
		}
	}()
	outLocation := fmt.Sprintf("%s/%s", p.outputPathPrefix, video.Location)
	if err = p.ProcessFileFromReader(ctx, rc, outLocation); err != nil {
		return fmt.Errorf("error processing file: %w", err)
	}
	return nil
}
