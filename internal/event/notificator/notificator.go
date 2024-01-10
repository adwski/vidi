package notificator

import (
	"context"
	"sync"

	"github.com/adwski/vidi/internal/api/video/client"
	"github.com/adwski/vidi/internal/event"
	"go.uber.org/zap"
)

const (
	defaultEventChannelLen = 100
)

type Notificator struct {
	c      *client.Client
	logger *zap.Logger
	evCh   chan *event.Event
}

type Config struct {
	Logger        *zap.Logger
	VideoAPIURL   string
	VideoAPIToken string
}

func New(cfg *Config) *Notificator {
	logger := cfg.Logger.With(zap.String("component", "notificator"))
	return &Notificator{
		logger: logger,
		evCh:   make(chan *event.Event, defaultEventChannelLen),
		c: client.New(&client.Config{
			Logger:   logger,
			Endpoint: cfg.VideoAPIURL,
			Token:    cfg.VideoAPIToken,
		})}
}

func (n *Notificator) Send(ev *event.Event) {
	n.evCh <- ev
}

func (n *Notificator) Run(ctx context.Context, wg *sync.WaitGroup) {
	n.logger.Info("started")
	defer wg.Done()
Loop:
	for {
		select {
		case <-ctx.Done():
			close(n.evCh)
			break Loop
		case ev := <-n.evCh:
			n.processEvent(ev)
		}
	}
	n.logger.Info("shutting down")
	for ev := range n.evCh {
		n.processEvent(ev)
	}
	n.logger.Info("stopped")
}

func (n *Notificator) processEvent(ev *event.Event) {
	n.logger.Debug("processing event", zap.Any("event", ev))
	var err error
	switch ev.Kind {
	case event.KindUpdateStatus:
		err = n.c.UpdateVideoStatus(ev.Video.ID, ev.Video.Status.String())
	case event.KindUpdateLocation:
		err = n.c.UpdateVideoLocation(ev.Video.ID, ev.Video.Location)
	case event.KindUpdateStatusAndLocation:
		err = n.c.UpdateVideo(ev.Video.ID, ev.Video.Status.String(), ev.Video.Location)
	default:
		n.logger.Error("unknown event kind", zap.Int("kind", ev.Kind))
		return
	}
	if err != nil {
		n.logger.Error("error while processing event", zap.Any("event", ev), zap.Error(err))
	}
	n.logger.Debug("event processed", zap.Any("event", ev))
}
