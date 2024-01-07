// Package processor contains mp4 file processor which can only do segmentation for now.
package processor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"

	mp4ff "github.com/Eyevinn/mp4ff/mp4"
)

const (
	segmentSuffixInit = "init.mp4"
	segmentSuffix     = ".m4s"
	mpdSuffix         = "manifest.mpd"
)

type MediaStore interface {
	Put(ctx context.Context, name string, r io.Reader, size int64) error
}

type Processor struct {
	logger *zap.Logger
	s      MediaStore
}

func NewProcessor(logger *zap.Logger, store MediaStore) *Processor {
	return &Processor{
		logger: logger.With(zap.String("component", "processor")),
		s:      store,
	}
}

// Process segments mp4 file provided as reader using specified segment duration
// and writes resulting segments to segment writer.
// It also generates MPD schema.
func (p *Processor) Process(ctx context.Context, r io.Reader, location string, segDuration time.Duration) error {
	p.logger.Info("mp4 processing started")
	mF, err := mp4ff.DecodeFile(r)
	if err != nil {
		return fmt.Errorf("cannot decode mp4 from reader: %w", err)
	}
	p.logger.Debug("mp4 decoded")

	tracks, timescale, totalDuration, errS := p.segmentFile(ctx, mF, segDuration, location)
	if errS != nil {
		return errS
	}
	if err = p.generateMPD(ctx, tracks, location, timescale, totalDuration, segDuration); err != nil {
		return err
	}

	p.logger.Info("mp4 file processed successfully")
	return nil
}

func (p *Processor) storeBytes(ctx context.Context, name string, artifact []byte) error {
	if err := p.s.Put(ctx, name, bytes.NewReader(artifact), int64(len(artifact))); err != nil {
		return fmt.Errorf("cannot write byte artifact: %w", err)
	}
	return nil
}

func (p *Processor) storeBox(ctx context.Context, name string, box mp4ff.BoxStructure, size uint64) error {
	var (
		errP, errE, errW, errR error
		r, w                   = io.Pipe()
		done                   = make(chan struct{})
	)
	go func() {
		errP = p.s.Put(ctx, name, r, int64(size))
		done <- struct{}{}
	}()
	go func() {
		errE = box.Encode(w)
		errW = w.Close() // this guarantees EOF in pipeReader
		done <- struct{}{}
	}()
	// a-ron
	<-done
	<-done
	if errP != nil {
		return fmt.Errorf("cannot put mp4 box into media store: %w", errP)
	}
	if errE != nil {
		return fmt.Errorf("cannot encode mp4 box: %w", errE)
	}
	if errW != nil {
		return fmt.Errorf("error closing encoder's writer: %w", errW)
	}
	if errR = r.Close(); errR != nil {
		return fmt.Errorf("error closing store reader: %w", errR)
	}
	return nil
}

func segmentName(track *mp4ff.TrakBox) string {
	return fmt.Sprintf("%s%d",
		track.Mdia.Hdlr.HandlerType, track.Tkhd.TrackID)
}
