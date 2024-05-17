package notificator

import (
	"bytes"
	"context"
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"

	"github.com/adwski/vidi/internal/event"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNotificator_RunAndStop(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	n := Notificator{
		logger: logger,
		evCh:   make(chan *event.Event, 10),
	}

	n.evCh <- &event.Event{Kind: 1000}
	n.evCh <- &event.Event{Kind: 1000}
	n.evCh <- &event.Event{Kind: 1000}

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go n.Run(ctx, wg, make(chan error))

	cancel()
	wg.Wait()
}

func TestNotificator_processUnknown(t *testing.T) {
	logBuf := bytes.NewBuffer([]byte{})
	logger := newLogger(logBuf)

	n := Notificator{
		logger: logger,
		evCh:   make(chan *event.Event),
	}

	n.processEvent(context.TODO(), &event.Event{
		Kind: 1000,
	})
	assert.Contains(t, logBuf.String(), "unknown event kind")
}

func newLogger(w io.Writer) *zap.Logger {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	return zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), zapcore.AddSync(w), zapcore.DebugLevel))
}
