package processor

import (
	"context"
	"fmt"
	"github.com/adwski/vidi/internal/mp4/meta"
	"io"

	mp4ff "github.com/Eyevinn/mp4ff/mp4"
	"github.com/adwski/vidi/internal/mp4/segmenter"
)

// ProcessFileFromReader segments mp4 file provided as reader using specified segment duration
// and writes resulting segments to segment writer.
// It also generates StaticMPD schema.
func (p *Processor) ProcessFileFromReader(ctx context.Context, rs io.ReadSeeker, location string) (*meta.Meta, error) {
	p.logger.Info("mp4 processing started")
	// Decoding in lazy mode.
	// Lazy mode will decode everything but will skip samples data in mdat.
	// Segmenter will read samples data directly from reader when necessary.
	mF, err := mp4ff.DecodeFile(rs, mp4ff.WithDecodeMode(mp4ff.DecModeLazyMdat))
	if err != nil {
		return nil, fmt.Errorf("cannot lazy decode mp4 from reader: %w", err)
	}
	p.logger.Debug("mp4 decoded")

	tracks, timescale, totalDuration, errS := segmenter.NewSegmenter(
		p.logger,
		rs,
		p.segmentDuration,
		func(ctx context.Context, name string, box mp4ff.BoxStructure, size uint64) error {
			return p.storeBox(ctx, fmt.Sprintf("%s/%s", location, name), box, size)
		}).SegmentMP4(ctx, mF)
	if errS != nil {
		return nil, fmt.Errorf("cannot segment mp4 file: %w", errS)
	}

	playbackMeta, err := p.generatePlaybackMeta(tracks, timescale, totalDuration)
	if err != nil {
		return nil, fmt.Errorf("cannot generate playback meta: %w", err)
	}
	p.logger.Info("mp4 file processed successfully")
	return playbackMeta, nil
}

func (p *Processor) storeBox(ctx context.Context, name string, box mp4ff.BoxStructure, size uint64) error {
	var (
		errP, errE, errW, errR error
		r, w                   = io.Pipe()
		done                   = make(chan struct{})
	)
	go func() {
		errP = p.st.Put(ctx, name, r, int64(size))
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
