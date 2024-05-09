// Package segmentation contains high level functions for progressive mp4 file segmentation
// The code is mostly based on https://github.com/Eyevinn/mp4ff/tree/master/examples/segmenter
package segmentation

import (
	"fmt"
	"io"
	"time"

	"github.com/Eyevinn/mp4ff/mp4"
)

const (
	millisecondsInSecond = 1000
)

// Point represents sample point in media track
// that will be used for segmentation.
type Point struct {
	sampleNum        uint32
	decodeTime       uint64
	presentationTime uint64
}

func GetFirstVideoTrackParams(m *mp4.File) (track *mp4.TrakBox, timescale uint32, duration uint64, err error) {
	for _, t := range m.Moov.Traks {
		if t.Mdia.Hdlr.HandlerType == "vide" {
			track = t
			break
		}
	}
	if track == nil {
		err = fmt.Errorf("no video track")
		return
	}
	timescale = track.Mdia.Mdhd.Timescale
	duration = track.Mdia.Mdhd.Duration
	return
}

// MakePoints creates segmentation points to split progressive mp4 file
// according to specified segment duration. It uses first video track
// as reference track for time calculations.
func MakePoints(track *mp4.TrakBox, timescale uint32, segmentDuration time.Duration) ([]Point, error) {
	var (
		segmentStep = uint32(uint64(segmentDuration.Milliseconds()) * uint64(timescale) / millisecondsInSecond)

		// https://youtu.be/CLvR9FVYwWs?t=840 (timing organisation)
		// https://youtu.be/CLvR9FVYwWs?t=922 (timelines)
		// ISO IEC 14496-12 8.5 Sample Tables
		stts = track.Mdia.Minf.Stbl.Stts // Time-to-sample box (decoding)
		ctts = track.Mdia.Minf.Stbl.Ctts // Time-to-sample box (composition)
		stss = track.Mdia.Minf.Stbl.Stss // Sync sample table

		nextSegmentStart uint32 // Next segment time mark

		// Allocate segmentation points array
		segmentationPoints = make([]Point, 0, stss.EntryCount())
	)

	for _, sampleNumber := range stss.SampleNumber {
		// Get decode time of the sample
		decodeTime, _ := stts.GetDecodeTime(sampleNumber)

		// Determine presentation time
		presentationTime := int64(decodeTime)
		if ctts != nil {
			// Correct by composition offset
			presentationTime += int64(ctts.GetCompositionTimeOffset(sampleNumber))
		}

		if presentationTime >= int64(nextSegmentStart) {
			// Time mark for next segmentation point is reached
			// Create it
			segmentationPoints = append(segmentationPoints,
				Point{
					sampleNum:        sampleNumber,
					decodeTime:       decodeTime,
					presentationTime: uint64(presentationTime),
				})
			// Update time mark
			nextSegmentStart += segmentStep
		}
	}
	return segmentationPoints, nil
}

// Interval represents segment by its start and end samples (inclusive).
type Interval struct {
	sampleStart uint32
	sampleEnd   uint32
}

// MakeIntervals returns sample intervals of track for specified segmentation points.
func MakeIntervals(timescale uint32, points []Point, track *mp4.TrakBox) ([]Interval, error) {
	var (
		startSampleNr     uint32 = 1
		nextStartSampleNr uint32 = 0
		endSampleNr       uint32
		err               error

		samplesCount    = track.Mdia.Minf.Stbl.Stsz.SampleNumber
		segmentCount    = len(points)
		sampleIntervals = make([]Interval, segmentCount)
	)

	for i := range points {
		if nextStartSampleNr != 0 {
			startSampleNr = nextStartSampleNr
		}
		if i == segmentCount-1 {
			endSampleNr = samplesCount - 1
		} else {
			nextSyncStart := points[i+1].decodeTime
			nextStartTime := nextSyncStart * uint64(track.Mdia.Mdhd.Timescale) / uint64(timescale)
			if nextStartSampleNr, err = track.Mdia.Minf.Stbl.Stts.GetSampleNrAtTime(nextStartTime); err != nil {
				return nil, fmt.Errorf("cannot sample number by time")
			}
			endSampleNr = nextStartSampleNr - 1
		}
		sampleIntervals[i] = Interval{
			sampleStart: startSampleNr,
			sampleEnd:   endSampleNr,
		}
	}
	return sampleIntervals, nil
}

// GetSamplesData retrieves media data for specified sample interval.
func GetSamplesData(mdat *mp4.MdatBox, stbl *mp4.StblBox, interval Interval, rs io.ReadSeeker) ([]mp4.FullSample, error) {
	if stbl.Stco == nil {
		return nil, fmt.Errorf("stco box is not present, co64 present: %v", stbl.Co64 != nil)
	}
	samples := make([]mp4.FullSample, 0, interval.sampleEnd-interval.sampleStart+1)
	payloadStart := mdat.PayloadAbsoluteOffset()

	for sampleNum := interval.sampleStart; sampleNum <= interval.sampleEnd; sampleNum++ {
		chunkNr, sampleNrAtChunkStart, err := stbl.Stsc.ChunkNrFromSampleNr(int(sampleNum))
		if err != nil {
			return nil, fmt.Errorf("cannot get chunk num for sample %d: %w", sampleNum, err)
		}
		if chunkNr > len(stbl.Stco.ChunkOffset) {
			return nil, fmt.Errorf("chunk number (%d) is greater than chunk offsets length (%d)",
				chunkNr, len(stbl.Stco.ChunkOffset))
		}
		offset := int64(stbl.Stco.ChunkOffset[chunkNr-1])
		for sNr := sampleNrAtChunkStart; sNr < int(sampleNum); sNr++ {
			offset += int64(stbl.Stsz.GetSampleSize(sNr))
		}
		size := stbl.Stsz.GetSampleSize(int(sampleNum))
		decTime, dur := stbl.Stts.GetDecodeTime(sampleNum)
		var cto int32 = 0
		if stbl.Ctts != nil {
			cto = stbl.Ctts.GetCompositionTimeOffset(sampleNum)
		}
		var sampleData []byte
		if mdat.GetLazyDataSize() > 0 {
			if rs == nil {
				return nil, fmt.Errorf("mdat decoded in lazy mode, but mdat reader is nil")
			}
			_, err = rs.Seek(offset, io.SeekStart)
			if err != nil {
				return nil, err
			}
			sampleData = make([]byte, size)
			_, err = io.ReadFull(rs, sampleData)
			if err != nil {
				return nil, err
			}
		} else {
			offsetInMdatData := uint64(offset) - payloadStart
			sampleData = mdat.Data[offsetInMdatData : offsetInMdatData+uint64(size)]
		}
		samples = append(samples, mp4.FullSample{
			Sample: mp4.Sample{
				Flags:                 translateSampleFlagsForFragment(stbl, sampleNum),
				Size:                  size,
				Dur:                   dur,
				CompositionTimeOffset: cto,
			},
			DecodeTime: decTime,
			Data:       sampleData,
		})
	}
	return samples, nil
}

// translateSampleFlagsForFragment - translate sample flags from stss and sdtp to what is needed in fragment
// Copied from mp4ff/examples/segmenter as is.
func translateSampleFlagsForFragment(stbl *mp4.StblBox, sampleNr uint32) (flags uint32) {
	var sampleFlags mp4.SampleFlags
	if stbl.Stss != nil {
		isSync := stbl.Stss.IsSyncSample(sampleNr)
		sampleFlags.SampleIsNonSync = !isSync
		if isSync {
			sampleFlags.SampleDependsOn = 2
		}
	}
	if stbl.Sdtp != nil {
		entry := stbl.Sdtp.Entries[sampleNr-1]
		sampleFlags.IsLeading = entry.IsLeading()
		sampleFlags.SampleDependsOn = entry.SampleDependsOn()
		sampleFlags.SampleHasRedundancy = entry.SampleHasRedundancy()
		sampleFlags.SampleIsDependedOn = entry.SampleIsDependedOn()
	}
	return sampleFlags.Encode()
}

// CreateInitForTrack creates initialization segment for specified track.
func CreateInitForTrack(track *mp4.TrakBox, timescale uint32, duration uint64) (*mp4.InitSegment, *mp4.TrakBox, error) {
	init := mp4.CreateEmptyInit()
	init.Moov.Mvhd.Timescale = timescale
	init.Moov.Mvex.AddChild(&mp4.MehdBox{
		FragmentDuration: int64(duration),
	})
	init.AddEmptyTrack(track.Mdia.Mdhd.Timescale, track.Mdia.Hdlr.HandlerType, track.Mdia.Mdhd.GetLanguage())
	var (
		outTrack  = init.Moov.Trak
		inStsd    = track.Mdia.Minf.Stbl.Stsd
		outStsd   = outTrack.Mdia.Minf.Stbl.Stsd
		trackType = track.Mdia.Hdlr.HandlerType
	)
	switch trackType {
	case "soun":
		switch {
		case inStsd.Mp4a != nil:
			outStsd.AddChild(inStsd.Mp4a)
		case inStsd.AC3 != nil:
			outStsd.AddChild(inStsd.AC3)
		case inStsd.EC3 != nil:
			outStsd.AddChild(inStsd.EC3)
		}
	case "vide":
		if inStsd.AvcX != nil {
			outStsd.AddChild(inStsd.AvcX)
		} else if inStsd.HvcX != nil {
			outStsd.AddChild(inStsd.HvcX)
		}
	default:
		return nil, nil, fmt.Errorf("unsupported track type: %s", trackType)
	}
	return init, outTrack, nil
}

// CreateSegment creates media segment with provided media data.
func CreateSegment(segNum int, trackID uint32, samplesData []mp4.FullSample) (*mp4.MediaSegment, error) {
	// Create fragment
	frag, errFr := mp4.CreateFragment(uint32(segNum), trackID)
	if errFr != nil {
		return nil, fmt.Errorf("cannot create fragment for track %d and segment %d: %w",
			trackID, segNum, errFr)
	}
	for _, sample := range samplesData {
		if err := frag.AddFullSampleToTrack(sample, trackID); err != nil {
			return nil, fmt.Errorf("cannot add sample data to fragment: %w", err)
		}
	}

	// Create segment
	seg := mp4.NewMediaSegment()
	seg.AddFragment(frag)
	return seg, nil
}
