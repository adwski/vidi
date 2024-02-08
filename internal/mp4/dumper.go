package mp4

import (
	"fmt"
	"io"
	"os"
	"time"

	mp4ff "github.com/Eyevinn/mp4ff/mp4"
	"github.com/adwski/vidi/internal/mp4/meta"
	"github.com/adwski/vidi/internal/mp4/segmentation"
)

const (
	defaultSegmentDuration = time.Second
)

// Dump prints out codec info and segmentation patter for mp4 file.
func Dump(path string, segDuration time.Duration) {
	dump(os.Stdout, path, segDuration)
}
func dump(w io.Writer, path string, segDuration time.Duration) {
	segmentDuration := segDuration
	if segDuration < defaultSegmentDuration {
		segmentDuration = defaultSegmentDuration
	}
	mF, err := mp4ff.ReadMP4File(path)
	if err != nil {
		fmt.Printf("cannot open mp4 file: %v\n", err)
		return
	}

	printW(w, "ftyp: %s\n", mF.Ftyp)
	printW(w, "segmented: %v\n", mF.IsFragmented())

	vTrack, timescale, totalDuration, errV := segmentation.GetFirstVideoTrackParams(mF)
	if errV != nil {
		fmt.Printf("cannot get first video track: %v\n", errV)
		return
	}
	printW(w, "timescale: %d units per second\n", timescale)
	printW(w, "duration: %v\n", time.Duration(totalDuration/uint64(timescale))*time.Second)
	printW(w, "\nsegmentation info:\n")

	segmentPoints, errSP := segmentation.MakePoints(vTrack, timescale, segmentDuration)
	printW(w, "segment points with %v duration (err: %v): %v\n", segmentDuration, errSP, segmentPoints)
	if errSP != nil {
		return
	}

	var (
		videoTrack bool
		audioTrack bool
	)
	for _, track := range mF.Moov.Traks {
		printW(w, "TrackID: %v, type: %v, sampleCount: %v\n",
			track.Tkhd.TrackID,
			track.Mdia.Hdlr.HandlerType,
			track.Mdia.Minf.Stbl.Stts.SampleCount)

		codec, errC := meta.NewCodecFromSTSD(track.Mdia.Minf.Stbl.Stsd)
		printW(w, "Codec info: %v (err: %v)\n", codec, errC)
		if errC == nil {
			switch track.Mdia.Hdlr.HandlerType {
			case "vide":
				videoTrack = true
			case "soun":
				audioTrack = true
			}
		}

		sI, errSI := segmentation.MakeIntervals(timescale, segmentPoints, track)
		printW(w, "Segment intervals (err: %v): %v\n", errSI, sI)
	}

	if videoTrack && audioTrack {
		printW(w, "\nCodecs are supported!\n")
	} else {
		printW(w, "\nSome codec is not yet supported!\n")
	}
}

func printW(w io.Writer, format string, a ...any) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		panic(err)
	}
}
