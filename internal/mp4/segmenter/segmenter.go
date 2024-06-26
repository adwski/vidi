package segmenter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	mp4ff "github.com/Eyevinn/mp4ff/mp4"
	"github.com/adwski/vidi/internal/mp4"
	"github.com/adwski/vidi/internal/mp4/segmentation"
	"go.uber.org/zap"
)

type BoxStoreFunc func(context.Context, string, mp4ff.BoxStructure, uint64) error

// Segmenter segments progressive mp4 according to predefined segment duration.
// Resulting segments are passed to boxStoreFunc, and it is up to user to define how to store them.
//
// Segmentation flow:
// 1) Make segmentation time-points using first video track
// 2) Make sample intervals that represent future segments
// 3) Create init segments for each supported track
// 4) Create data segments for each supported track
//
// This flow uses high-level functions, implemented in segmentation package.
//
// Configured segment duration should be treated like 'preference'.
// Segmenter can increase it if necessary in order to make segments with equal sizes.
//
// Tracks get new numbers since they will be in separate files.
// For example, if input file has video track with ID 0 and audio with ID 1,
// then in resulting video segments will be single track with ID 0,
// and in resulting audio segments will be single track also with ID 0.
//
// Approach is taken from https://github.com/Eyevinn/mp4ff/tree/master/examples/segmenter.
type Segmenter struct {
	logger          *zap.Logger
	boxStoreFunc    BoxStoreFunc
	mdatRS          io.ReadSeeker
	segmentDuration time.Duration
}

func NewSegmenter(
	logger *zap.Logger,
	mdatRS io.ReadSeeker,
	segDuration time.Duration,
	boxStoreFunc BoxStoreFunc,
) *Segmenter {
	return &Segmenter{
		logger:          logger,
		mdatRS:          mdatRS,
		boxStoreFunc:    boxStoreFunc,
		segmentDuration: segDuration,
	}
}

func (s *Segmenter) GetSegmentDuration() time.Duration { return s.segmentDuration }

func (s *Segmenter) SegmentMP4(
	ctx context.Context,
	mF *mp4ff.File,
) (map[uint32]*mp4ff.TrakBox, uint32, uint64, error) {
	track, timescale, totalDuration, err := segmentation.GetFirstVideoTrackParams(mF)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("cannot get first video track: %w", err)
	}
	s.logger.Debug("main video track params retrieved",
		zap.String("type", track.Mdia.Hdlr.HandlerType),
		zap.Uint32("timescale", timescale),
		zap.Uint64("totalDuration", totalDuration))

	updSegDuration, segPoints, errS := segmentation.MakePoints(track, timescale, s.segmentDuration)
	if updSegDuration > 0 {
		s.logger.Debug("updating segment duration", zap.Duration("new", updSegDuration))
		s.segmentDuration = updSegDuration
	}
	if errS != nil {
		return nil, 0, 0, fmt.Errorf("cannot make segmentation points: %w", err)
	}
	s.logger.Debug("segmentation points calculated",
		zap.Duration("segmentDuration", s.segmentDuration),
		zap.Int("count", len(segPoints)))

	suitableTracks, errTr := s.getSuitableTracks(mF)
	if errTr != nil {
		return nil, 0, 0, fmt.Errorf("cannot find suitable tracks: %w", errTr)
	}
	s.logger.Info("found tracks", zap.Int("tracks", len(suitableTracks)))

	segTracks, errSeg := s.makeAndWriteInitSegments(ctx, suitableTracks, timescale, totalDuration)
	if errSeg != nil {
		return nil, 0, 0, fmt.Errorf("cannot write init segments: %w", errSeg)
	}
	s.logger.Debug("init segments generated",
		zap.Int("segmentedTracks", len(segTracks)))

	for _, tr := range suitableTracks {
		segments, errIn := segmentation.MakeIntervals(timescale, segPoints, tr)
		if errIn != nil {
			return nil, 0, 0, fmt.Errorf("cannot make segment intervals: %w", err)
		}
		s.logger.Debug("track segments generated",
			zap.String("type", tr.Mdia.Hdlr.HandlerType),
			zap.Int("segments", len(segments)))

		if err = s.makeAndWriteSegments(ctx, segments, tr, segTracks, mF.Mdat); err != nil {
			return nil, 0, 0, fmt.Errorf("error during segment processing: %w", err)
		}
		s.logger.Debug("track segments sent to storage",
			zap.String("type", tr.Mdia.Hdlr.HandlerType))
	}
	s.logger.Info("mp4 segmented successfully")
	return segTracks, timescale, totalDuration, nil
}

func (s *Segmenter) getSuitableTracks(m *mp4ff.File) ([]*mp4ff.TrakBox, error) {
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
			s.logger.Warn("got unknown track",
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
func (s *Segmenter) makeAndWriteSegments(
	ctx context.Context,
	segments []segmentation.Interval,
	track *mp4ff.TrakBox,
	segTracks map[uint32]*mp4ff.TrakBox,
	mdat *mp4ff.MdatBox,
) error {
	for i, segInterval := range segments {
		segNum := i + 1
		segTrackID := segTracks[track.Tkhd.TrackID].Tkhd.TrackID
		// Get segments data for segment
		samplesData, err := segmentation.GetSamplesData(mdat, track.Mdia.Minf.Stbl, segInterval, s.mdatRS)
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
		name := fmt.Sprintf("%s_%d%s", mp4.SegmentName(segTracks[track.Tkhd.TrackID]), segNum, mp4.SegmentSuffix)
		if err = s.boxStoreFunc(ctx, name, seg, seg.Size()); err != nil {
			return err
		}
	}
	return nil
}

func (s *Segmenter) makeAndWriteInitSegments(
	ctx context.Context,
	tracks []*mp4ff.TrakBox,
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
		name := fmt.Sprintf("%s_%s", mp4.SegmentName(segTrack), mp4.SegmentSuffixInit)
		if err = s.boxStoreFunc(ctx, name, init, init.Size()); err != nil {
			return nil, err
		}
		segTracks[track.Tkhd.TrackID] = segTrack
	}
	return segTracks, nil
}
