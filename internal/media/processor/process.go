package processor

import (
	"bytes"
	"context"
	"fmt"
	"io"

	mp4ff "github.com/Eyevinn/mp4ff/mp4"
	"github.com/adwski/vidi/internal/mp4"
	"github.com/adwski/vidi/internal/mp4/segmenter"
)

// ProcessFileFromReader segments mp4 file provided as reader using specified segment duration
// and writes resulting segments to segment writer.
// It also generates MPD schema.
func (p *Processor) ProcessFileFromReader(ctx context.Context, r io.Reader, location string) error {
	p.logger.Info("mp4 processing started")
	mF, err := mp4ff.DecodeFile(r)
	if err != nil {
		return fmt.Errorf("cannot decode mp4 from reader: %w", err)
	}
	p.logger.Debug("mp4 decoded")

	tracks, timescale, totalDuration, errS := segmenter.NewSegmenter(
		p.logger,
		p.segmentDuration,
		func(ctx context.Context, name string, box mp4ff.BoxStructure, size uint64) error {
			return p.storeBox(ctx, fmt.Sprintf("%s/%s", location, name), box, size)
		}).SegmentMP4(ctx, mF)
	if errS != nil {
		return fmt.Errorf("cannot segment mp4 file: %w", errS)
	}

	bMPD, errMPD := p.constructMetadataAndGenerateMPD(tracks, timescale, totalDuration)
	if errMPD != nil {
		return fmt.Errorf("cannot generate MPD: %w", errMPD)
	}
	p.logger.Debug("mpd generated successfully")
	if err = p.storeBytes(ctx, fmt.Sprintf("%s/%s", location, mp4.MPDSuffix), bMPD); err != nil {
		return err
	}

	p.logger.Info("mp4 file processed successfully")
	return nil
}

func (p *Processor) storeBytes(ctx context.Context, name string, artifact []byte) error {
	if err := p.st.Put(ctx, name, bytes.NewReader(artifact), int64(len(artifact))); err != nil {
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
