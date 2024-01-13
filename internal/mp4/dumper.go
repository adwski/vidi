package mp4

import (
	"fmt"
	"time"

	mp4ff "github.com/Eyevinn/mp4ff/mp4"
	"github.com/adwski/vidi/internal/mp4/meta"
	"github.com/adwski/vidi/internal/mp4/segmentation"
)

const (
	defaultSegmentDuration = time.Second
)

func Dump(path string, segDuration time.Duration) {
	segmentDuration := segDuration
	if segDuration < defaultSegmentDuration {
		segmentDuration = defaultSegmentDuration
	}
	mF, err := mp4ff.ReadMP4File(path)
	if err != nil {
		fmt.Printf("cannot open mp4 file: %v\n", err)
		return
	}

	fmt.Printf("ftyp: %s\n", mF.Ftyp)
	fmt.Printf("segmented: %v\n", mF.IsFragmented())

	vTrack, timescale, totalDuration, errV := segmentation.GetFirstVideoTrackParams(mF)
	if errV != nil {
		fmt.Printf("cannot get first video track: %v\n", errV)
		return
	}
	fmt.Printf("timescale: %d units per second\n", timescale)
	fmt.Printf("duration: %v\n", time.Duration(totalDuration/uint64(timescale))*time.Second)
	fmt.Println("\nsegmentation info:")

	segmentPoints, errSP := segmentation.MakePoints(vTrack, timescale, segmentDuration)
	fmt.Printf("segment points with %v duration (err: %v): %v\n", segmentDuration, errSP, segmentPoints)
	if errSP != nil {
		return
	}

	var (
		videoTrack bool
		audioTrack bool
	)
	for _, track := range mF.Moov.Traks {
		fmt.Printf("TrackID: %v, type: %v, sampleCount: %v\n",
			track.Tkhd.TrackID,
			track.Mdia.Hdlr.HandlerType,
			track.Mdia.Minf.Stbl.Stts.SampleCount)

		codec, errC := meta.NewCodecFromSTSD(track.Mdia.Minf.Stbl.Stsd)
		fmt.Printf("Codec info: %v (err: %v)\n", codec, errC)
		if errC == nil {
			switch track.Mdia.Hdlr.HandlerType {
			case "vide":
				videoTrack = true
			case "soun":
				audioTrack = true
			}
		}

		sI, errSI := segmentation.MakeIntervals(timescale, segmentPoints, track)
		fmt.Printf("Segment intervals (err: %v): %v\n", errSI, sI)
	}

	if videoTrack && audioTrack {
		fmt.Println("\nCodecs are supported!")
	} else {
		fmt.Println("\nSome codec is not yet supported!")
	}
}
