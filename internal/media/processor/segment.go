package processor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	mp4ff "github.com/Eyevinn/mp4ff/mp4"
	"github.com/adwski/vidi/internal/mp4/segmentation"
)

func (p *Processor) segmentFile(
	ctx context.Context,
	mF *mp4ff.File,
	segDuration time.Duration,
	location string,
) (map[uint32]*mp4ff.TrakBox, uint32, uint64, error) {
	track, timescale, totalDuration, err := segmentation.GetFirstVideoTrackParams(mF)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("cannot get first video track: %w", err)
	}
	p.logger.Debug("main video track params retrieved",
		zap.String("type", track.Mdia.Hdlr.HandlerType),
		zap.Uint32("timescale", timescale),
		zap.Uint64("totalDuration", totalDuration))

	segPoints, errS := segmentation.MakePoints(track, timescale, segDuration)
	if errS != nil {
		return nil, 0, 0, fmt.Errorf("cannot make segmentation points: %w", err)
	}
	p.logger.Debug("segmentation points calculated",
		zap.Int("count", len(segPoints)))

	suitableTracks, errTr := p.getSuitableTracks(mF)
	if errTr != nil {
		return nil, 0, 0, fmt.Errorf("cannot find suitable tracks: %w", errTr)
	}
	p.logger.Info("found tracks", zap.Int("tracks", len(suitableTracks)))

	segTracks, errSeg := p.makeAndWriteInitSegments(ctx, suitableTracks, location, timescale, totalDuration)
	if errSeg != nil {
		return nil, 0, 0, fmt.Errorf("cannot write init segments: %w", errSeg)
	}
	p.logger.Debug("init segments generated",
		zap.Int("segmentedTracks", len(segTracks)))

	for _, tr := range suitableTracks {
		segments, errIn := segmentation.MakeIntervals(timescale, segPoints, tr)
		if errIn != nil {
			return nil, 0, 0, fmt.Errorf("cannot make segment intervals: %w", err)
		}
		p.logger.Debug("track segments generated",
			zap.String("type", tr.Mdia.Hdlr.HandlerType),
			zap.Int("segments", len(segments)))

		if err = p.makeAndWriteSegments(ctx, segments, tr, segTracks, mF.Mdat, location); err != nil {
			return nil, 0, 0, fmt.Errorf("error during segment processing: %w", err)
		}
		p.logger.Debug("track segments sent to storage",
			zap.String("type", tr.Mdia.Hdlr.HandlerType))
	}
	p.logger.Info("mp4 segmented successfully")
	return segTracks, timescale, totalDuration, nil
}

func (p *Processor) getSuitableTracks(m *mp4ff.File) ([]*mp4ff.TrakBox, error) {
	var (
		vide      bool
		outTracks = make([]*mp4ff.TrakBox, 0, len(m.Moov.Traks))
	)
	for _, track := range m.Moov.Traks {
		switch track.Mdia.Hdlr.HandlerType {
		case "vide":
			vide = true
		case "soun":
		default:
			p.logger.Warn("got unknown track",
				zap.String("type", track.Mdia.Hdlr.HandlerType))
			continue
		}
		outTracks = append(outTracks, track)
	}
	if len(outTracks) == 0 {
		return nil, errors.New("mp4 does not have video or audio tracks")
	}
	if !vide {
		return nil, errors.New("mp4 does not have video tracks")
	}
	return outTracks, nil
}
func (p *Processor) makeAndWriteSegments(
	ctx context.Context,
	segments []segmentation.Interval,
	track *mp4ff.TrakBox,
	segTracks map[uint32]*mp4ff.TrakBox,
	mdat *mp4ff.MdatBox,
	location string,
) error {
	for i, segInterval := range segments {
		segNum := i + 1
		segTrackID := segTracks[track.Tkhd.TrackID].Tkhd.TrackID
		// Get segments data for segment
		samplesData, err := segmentation.GetSamplesData(mdat, track.Mdia.Minf.Stbl, segInterval)
		if err != nil {
			return fmt.Errorf("cannot get samples data: %w", err)
		}
		if len(samplesData) == 0 {
			// no samples for track, but looks like it might be ok
			continue
		}

		// Create mp4 segment with data
		seg, errS := segmentation.CreateSegment(segNum, segTrackID, samplesData)
		if errS != nil {
			return fmt.Errorf("cannot create media segment:%w", errS)
		}

		// Write segment
		name := fmt.Sprintf("%s%s_%d%s", location, segmentName(segTracks[track.Tkhd.TrackID]), segNum, segmentSuffix)
		if err = p.storeBox(ctx, name, seg, seg.Size()); err != nil {
			return err
		}
	}
	return nil
}

func (p *Processor) makeAndWriteInitSegments(
	ctx context.Context,
	tracks []*mp4ff.TrakBox,
	location string,
	timescale uint32,
	duration uint64,
) (map[uint32]*mp4ff.TrakBox, error) {
	var (
		segTracks = make(map[uint32]*mp4ff.TrakBox) // old track num -> new track
	)
	for _, track := range tracks {
		init, segTrack, err := segmentation.CreateInitForTrack(track, timescale, duration)
		if err != nil {
			return nil, fmt.Errorf("cannot create init track: %w", err)
		}
		name := fmt.Sprintf("%s%s_%s", location, segmentName(segTrack), segmentSuffixInit)
		if err = p.storeBox(ctx, name, init, init.Size()); err != nil {
			return nil, err
		}
		segTracks[track.Tkhd.TrackID] = segTrack
	}
	return segTracks, nil
}
