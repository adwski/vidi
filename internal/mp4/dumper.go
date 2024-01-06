package mp4

import (
	"fmt"
	"time"

	"github.com/Eyevinn/mp4ff/mp4"
	"github.com/adwski/vidi/internal/dash"
	"github.com/adwski/vidi/internal/mp4/segmentation"
)

const (
	segmentDuration = time.Second
)

func Dump(path string) {
	mF, err := mp4.ReadMP4File(path)
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

	segmentPoints, errSP := segmentation.MakePoints(vTrack, timescale, segmentDuration)
	fmt.Printf("segment points (err: %v) with %v duration: %v\n", errSP, segmentDuration, segmentPoints)
	if errSP != nil {
		return
	}

	for _, track := range mF.Moov.Traks {
		fmt.Printf("TrackID: %v, type: %v, sampleCount: %v\n",
			track.Tkhd.TrackID,
			track.Mdia.Hdlr.HandlerType,
			track.Mdia.Minf.Stbl.Stts.SampleCount)

		codec, errC := dash.NewCodecFromTrackSTSD(track.Mdia.Minf.Stbl.Stsd)
		fmt.Printf("Codec info: %v (err: %v)\n", codec, errC)

		sI, errSI := segmentation.MakeIntervals(timescale, segmentPoints, track)
		fmt.Printf("Segment intervals (err: %v): %v\n", errSI, sI)
	}
}
