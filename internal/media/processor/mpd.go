package processor

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	mp4ff "github.com/Eyevinn/mp4ff/mp4"
	"github.com/adwski/vidi/internal/dash"
)

func (p *Processor) generateMPD(
	ctx context.Context,
	tracks map[uint32]*mp4ff.TrakBox,
	location string,
	timescale uint32,
	totalDuration uint64,
	segmentDuration time.Duration,
) error {
	var (
		i          int
		dashTracks = make([]dash.Track, len(tracks))
	)
	for _, track := range tracks {
		mimeType, err := dash.GetMimeTypeFromMP4TrackHandlerType(track.Mdia.Hdlr.HandlerType)
		if err != nil {
			return fmt.Errorf("cannot get handler type: %w", err)
		}

		codec, errC := dash.NewCodecFromTrackSTSD(track.Mdia.Minf.Stbl.Stsd)
		if errC != nil {
			return fmt.Errorf("cannot get codec: %w", errC)
		}

		dashTracks[i] = dash.Track{
			Name:     segmentName(track),
			MimeType: mimeType,
			Codec:    codec,
			// TODO idk if bandwidth is mandatory for only one representation (probably not).
			//   For future, it probably could be estimated using avg bitrate
			//   (but not every file has Btrt box). Another way would be to statically define
			//   bandwidth if we re-encode file to support multiple representations.
			// Bandwidth: 0,
			Segment: &dash.SegmentConfig{
				Init:        segmentSuffixInit,
				StartNumber: 1,
				// TODO Last segment duration most probably will not be equal to segmentDuration
				//   Is this important? Should clients handle this on their side?
				Duration:  uint64(segmentDuration.Seconds() * float64(timescale)),
				Timescale: timescale,
			},
		}
		i++
	}

	dashCfg := &dash.MediaConfig{
		Duration: time.Duration(int64(totalDuration)/int64(timescale)) * time.Second,
		Tracks:   dashTracks,
	}
	b, err := dashCfg.GenerateMPD()
	if err != nil {
		return fmt.Errorf("cannot generate MPD: %w", err)
	}
	name := fmt.Sprintf("%s%s", location, mpdSuffix)
	if err = p.storeBytes(ctx, name, b); err != nil {
		return err
	}
	p.logger.Info("mpd generated successfully",
		zap.Duration("duration", dashCfg.Duration),
		zap.Duration("segmentDuration", segmentDuration),
		zap.String("name", name),
		zap.Any("tracks", dashCfg.TracksInfo()))
	return nil
}
