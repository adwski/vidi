package processor

import (
	"fmt"
	"time"

	mp4ff "github.com/Eyevinn/mp4ff/mp4"
	"github.com/adwski/vidi/internal/mp4"
	"github.com/adwski/vidi/internal/mp4/meta"
)

func (p *Processor) generatePlaybackMeta(
	tracks map[uint32]*mp4ff.TrakBox,
	timescale uint32,
	totalDuration uint64,
) (*meta.Meta, error) {
	var (
		i          int
		dashTracks = make([]meta.Track, len(tracks))
	)
	for _, track := range tracks {
		mimeType, err := getMimeTypeFromMP4TrackHandlerType(track.Mdia.Hdlr.HandlerType)
		if err != nil {
			return nil, fmt.Errorf("cannot get handler type: %w", err)
		}

		codec, errC := meta.NewCodecFromSTSD(track.Mdia.Minf.Stbl.Stsd)
		if errC != nil {
			return nil, fmt.Errorf("cannot get codec: %w", errC)
		}

		dashTracks[i] = meta.Track{
			Name:     mp4.SegmentName(track),
			MimeType: mimeType,
			Codec:    codec,
			// TODO idk if bandwidth is mandatory for only one representation (probably not).
			//   For future, it probably could be estimated using avg bitrate
			//   (but not every file has Btrt box). Another way would be to statically define
			//   bandwidth if we re-encode file to support multiple representations.
			// Bandwidth: 0,
			Segment: &meta.SegmentConfig{
				Init:        mp4.SegmentSuffixInit,
				StartNumber: 1,
				// TODO Last segment duration most probably will not be equal to segmentDuration
				//   Is this important? Should clients handle this on their side?
				Duration:  uint64(p.segmentDuration.Seconds() * float64(timescale)),
				Timescale: timescale,
			},
		}
		i++
	}
	return &meta.Meta{
		Duration: time.Duration(int64(totalDuration)/int64(timescale)) * time.Second,
		Tracks:   dashTracks,
	}, nil
}

/*
	b, err := metaCfg.StaticMPD()
	if err != nil {
		return nil, fmt.Errorf("cannot generate MPD: %w", err)
	}
	return b, nil
*/

func getMimeTypeFromMP4TrackHandlerType(handlerType string) (string, error) {
	switch handlerType {
	case "soun":
		return "audio/mp4", nil
	case "vide":
		return "video/mp4", nil
	}
	return "", fmt.Errorf("unknown mp4 track handler type: %s", handlerType)
}
