// Package processor contains mp4 file processor which can only do segmentation for now.
package processor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/adwski/vidi/internal/api/video/grpc/serviceside/pb"
	video "github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/event"
	"github.com/adwski/vidi/internal/event/notificator"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	defaultPartSize = 10 * 1024 * 1024
)

type MediaStore interface {
	Put(ctx context.Context, name string, r io.Reader, size int64) error
	Get(ctx context.Context, name string) (io.ReadSeekCloser, int64, error)
}

// Processor is worker-style app that polls videoapi for uploaded videos,
// processes them and notifies videoapi about results.
//
// Processing includes
// - segmentation
// - MPD generation
// Segments and static MPD are stored in MediaStore (s3).
type Processor struct {
	logger           *zap.Logger
	notificator      *notificator.Notificator
	videoAPI         pb.ServicesideapiClient
	authMD           metadata.MD
	st               MediaStore
	inputPathPrefix  string
	outputPathPrefix string
	segmentDuration  time.Duration
	videoCheckPeriod time.Duration
}

type Config struct {
	Logger           *zap.Logger
	Notificator      *notificator.Notificator
	Store            MediaStore
	VideoAPIEndpoint string
	VideoAPIToken    string
	InputPathPrefix  string
	OutputPathPrefix string
	SegmentDuration  time.Duration
	VideoCheckPeriod time.Duration
}

func New(cfg *Config) (*Processor, error) {
	logger := cfg.Logger.With(zap.String("component", "processor"))
	if cfg.VideoAPIEndpoint == "" {
		logger.Debug("running in local mode")
		return &Processor{
			logger:          cfg.Logger.With(zap.String("component", "processor")),
			st:              cfg.Store,
			segmentDuration: cfg.SegmentDuration,
		}, nil
	}
	cc, err := grpc.Dial(cfg.VideoAPIEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("cannot create vidi connection: %w", err)
	}
	return &Processor{
		logger:           cfg.Logger.With(zap.String("component", "processor")),
		st:               cfg.Store,
		notificator:      cfg.Notificator,
		segmentDuration:  cfg.SegmentDuration,
		videoCheckPeriod: cfg.VideoCheckPeriod,
		inputPathPrefix:  strings.TrimSuffix(cfg.InputPathPrefix, "/"),
		outputPathPrefix: strings.TrimSuffix(cfg.OutputPathPrefix, "/"),
		videoAPI:         pb.NewServicesideapiClient(cc),
		authMD:           metadata.Pairs("authorization", "bearer "+cfg.VideoAPIToken),
	}, nil
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

func (p *Processor) getUploadedVideos(ctx context.Context) ([]*pb.Video, error) {
	resp, err := p.videoAPI.GetVideosByStatus(metadata.NewOutgoingContext(ctx, p.authMD),
		&pb.GetByStatusRequest{Status: int32(video.StatusUploaded)})
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return nil, fmt.Errorf("unable to retrieve uploaded videos: %w", err)
		}
		return nil, nil
	}
	return resp.Videos, nil
}

func (p *Processor) checkAndProcessVideos(ctx context.Context) {
	p.logger.Debug("checking videos")

	videos, err := p.getUploadedVideos(ctx)
	if err != nil {
		p.logger.Error("cannot get videos from video API", zap.Error(err))
		return
	}
	if len(videos) == 0 {
		p.logger.Debug("no videos for processing")
		return
	}

	p.logger.Info("got videos for processing", zap.Int("count", len(videos)))

	for _, v := range videos {
		bMPD, err := p.processVideo(ctx, v)
		if err != nil {
			// TODO In the future we should distinguish between errors caused by video content
			//  and any other error. For example i/o errors are related to other failures
			//  and in such cases video processing could be retried later. (So we need retry mechanism).
			p.notificator.Send(&event.Event{
				VideoInfo: &event.VideoInfo{
					VideoID: v.Id,
					Status:  int(video.StatusError),
				},
				Kind: event.KindUpdateStatus,
			})
			p.logger.Error("error while processing video",
				zap.String("id", v.Id),
				zap.Error(err))
			continue
		}
		p.notificator.Send(&event.Event{
			VideoInfo: &event.VideoInfo{
				VideoID: v.Id,
				Meta:    bMPD,
			},
			Kind: event.KindVideoReady,
		})
		p.logger.Debug("video processed successfully",
			zap.String("id", v.Id))
	}
	p.logger.Debug("processing done")
}

func (p *Processor) processVideo(ctx context.Context, v *pb.Video) ([]byte, error) {
	p.logger.Debug("processing video",
		zap.String("id", v.Id),
		zap.String("location", v.Location),
		zap.Int("parts", len(v.Parts)),
		zap.Uint64("size", v.Size))
	switch {
	case len(v.Parts) == 0:
		return nil, errors.New("video has no parts")
	case v.Size == 0:
		return nil, errors.New("video has zero size")
	case defaultPartSize*uint64(len(v.Parts)-1) > v.Size || v.Size > defaultPartSize*uint64(len(v.Parts)):
		return nil, fmt.Errorf("incorrect parts amount(%d) for video size(%d)", len(v.Parts), v.Size)
	}
	mr := newMediaReader(
		p.st,
		fmt.Sprintf("%s/%s", p.inputPathPrefix, v.Location),
		uint(len(v.Parts)),
		v.Size,
		defaultPartSize)
	defer func() {
		if err := mr.Close(); err != nil {
			p.logger.Error("error closing media reader",
				zap.Error(err),
				zap.String("vid", v.Id))
		}
	}()
	outLocation := fmt.Sprintf("%s/%s", p.outputPathPrefix, v.Location)
	playbackMeta, err := p.ProcessFileFromReader(ctx, mr, outLocation)
	if err != nil {
		return nil, fmt.Errorf("error processing file: %w", err)
	}
	b, err := msgpack.Marshal(playbackMeta)
	if err != nil {
		return nil, fmt.Errorf("cannot marshall playback meta to msgpack: %w", err)
	}
	return b, nil
}
