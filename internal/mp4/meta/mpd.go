package meta

import (
	"fmt"
	"strconv"

	"github.com/Eyevinn/dash-mpd/mpd"
	"github.com/Eyevinn/dash-mpd/xml"
)

// StaticMPD generates media presentation description (MPD) corresponding to current state of Meta.
// Based on https://github.com/Eyevinn/dash-mpd/blob/main/examples/newmpd_test.go
// Refs: ISO/IEC 23009-1 4.3 DASH data model overview.
func (mc *Meta) StaticMPD() ([]byte, error) {
	// Create StaticMPD
	m := mpd.NewMPD(mpd.STATIC_TYPE)
	m.Profiles = mpd.PROFILE_ONDEMAND
	m.MediaPresentationDuration = mpd.Ptr(mpd.Duration(mc.Duration))

	// Create Period
	p := mpd.NewPeriod()
	p.Id = "p0"
	m.AppendPeriod(p)

	// Create adaptation sets
	for _, track := range mc.Tracks {
		p.AppendAdaptationSet(track.makeAdaptationSet())
	}

	// Marshall XML using patched encoding/xml
	out, err := xml.MarshalIndent(m, " ", "")
	if err != nil {
		return nil, fmt.Errorf("cannot marshal mpd: %w", err)
	}
	return out, nil
}

func (track *Track) makeAdaptationSet() *mpd.AdaptationSetType {
	// Create AdaptationSet
	as := mpd.NewAdaptationSet()
	as.MimeType = track.MimeType

	// Create SegmentTemplate
	st := mpd.NewSegmentTemplate()
	st.StartNumber = mpd.Ptr(uint32(track.Segment.StartNumber))
	st.Timescale = mpd.Ptr(track.Segment.Timescale)
	st.Duration = mpd.Ptr(uint32(track.Segment.Duration))
	st.Initialization = fmt.Sprintf("$RepresentationID$_%s", track.Segment.Init)
	st.Media = "$RepresentationID$_$Number$.m4s"
	as.SegmentTemplate = st

	// Create representation
	rep := mpd.NewRepresentation()
	rep.Id = track.Name
	rep.Codecs = track.Codec.Profile
	if track.Bandwidth != 0 {
		rep.Bandwidth = track.Bandwidth
	}
	if track.Codec.SampleRate != 0 {
		rep.AudioSamplingRate = mpd.Ptr(mpd.UIntVectorType(strconv.Itoa(int(track.Codec.SampleRate))))
	}
	as.AppendRepresentation(rep)

	return as
}
