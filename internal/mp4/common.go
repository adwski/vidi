package mp4

import (
	"fmt"

	mp4ff "github.com/Eyevinn/mp4ff/mp4"
)

const (
	SegmentSuffixInit = "init.mp4"
	SegmentSuffix     = ".m4s"
	MPDSuffix         = "manifest.mpd"
)

func SegmentName(track *mp4ff.TrakBox) string {
	return fmt.Sprintf("%s%d",
		track.Mdia.Hdlr.HandlerType, track.Tkhd.TrackID)
}
