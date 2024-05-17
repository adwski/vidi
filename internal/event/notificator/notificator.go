package notificator

import (
	"context"
	"fmt"
	"sync"

	"github.com/adwski/vidi/internal/api/video/grpc/serviceside/pb"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/event"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	defaultEventChannelLen = 100
)

// Notificator is asynchronous Video API notification service.
// It takes events and calls videoapi service-side API in separate goroutine.
//
// TODO In the future could be replaced with actual message queue.
type Notificator struct {
	c      pb.ServicesideapiClient
	authMD metadata.MD
	logger *zap.Logger
	evCh   chan *event.Event
}

type Config struct {
	Logger        *zap.Logger
	VideoAPIURL   string
	VideoAPIToken string
}

func New(cfg *Config) (*Notificator, error) {
	cc, err := grpc.Dial(cfg.VideoAPIURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("cannot create vidi connection: %w", err)
	}
	return &Notificator{
		authMD: metadata.Pairs("authorization", "bearer "+cfg.VideoAPIToken),
		logger: cfg.Logger.With(zap.String("component", "notificator")),
		evCh:   make(chan *event.Event, defaultEventChannelLen),
		c:      pb.NewServicesideapiClient(cc),
	}, nil
}

func (n *Notificator) Send(ev *event.Event) {
	n.evCh <- ev
}

func (n *Notificator) Run(ctx context.Context, wg *sync.WaitGroup, _ chan<- error) {
	n.logger.Info("started")
	defer wg.Done()
Loop:
	for {
		select {
		case <-ctx.Done():
			close(n.evCh)
			break Loop
		case ev := <-n.evCh:
			go n.processEvent(ctx, ev)
		}
	}
	n.logger.Info("stopping")
	for ev := range n.evCh {
		n.processEvent(ctx, ev)
	}
	n.logger.Info("stopped")
}

func (n *Notificator) processEvent(ctx context.Context, ev *event.Event) {
	n.logger.Debug("processing event", zap.Any("event", ev))

	switch ev.Kind {
	case event.KindVideoPartUploaded:
		if ev.PartInfo == nil {
			n.logger.Error("PartInfo is nil", zap.Int("kind", ev.Kind))
			return
		}
	case event.KindUpdateStatus, event.KindVideoReady:
		if ev.VideoInfo == nil {
			n.logger.Error("VideoInfo is nil", zap.Int("kind", ev.Kind))
			return
		}
	default:
		n.logger.Error("unknown event kind", zap.Int("kind", ev.Kind))
		return
	}

	var err error
	switch ev.Kind {
	case event.KindVideoPartUploaded:
		_, err = n.c.NotifyPartUpload(metadata.NewOutgoingContext(ctx, n.authMD), &pb.NotifyPartUploadRequest{
			Num:      uint32(ev.PartInfo.Num),
			VideoId:  ev.PartInfo.VideoID,
			Checksum: ev.PartInfo.Checksum,
		})
	case event.KindUpdateStatus:
		_, err = n.c.UpdateVideoStatus(metadata.NewOutgoingContext(ctx, n.authMD), &pb.UpdateVideoStatusRequest{
			Id:     ev.VideoInfo.VideoID,
			Status: int32(ev.VideoInfo.Status),
		})
	case event.KindVideoReady:
		_, err = n.c.UpdateVideo(metadata.NewOutgoingContext(ctx, n.authMD), &pb.UpdateVideoRequest{
			Id:           ev.VideoInfo.VideoID,
			Status:       int32(model.StatusReady),
			PlaybackMeta: ev.VideoInfo.Meta,
		})
	}

	if err != nil {
		n.logger.Error("error while processing event", zap.Any("event", ev), zap.Error(err))
	}
	n.logger.Debug("event processed", zap.Any("event", ev))
}
