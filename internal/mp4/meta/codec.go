package meta

import (
	"bytes"
	"fmt"

	"github.com/Eyevinn/mp4ff/aac"
	mp4ff "github.com/Eyevinn/mp4ff/mp4"
	"github.com/davecgh/go-spew/spew"
)

// Codec holds generic codec info of audio or video track.
type Codec struct {
	Profile    string
	SampleRate uint16
}

// NewCodecFromSTSD retrieves codec info from track's STSD box
// in the format suited for DASH MPD schema.
func NewCodecFromSTSD(stsd *mp4ff.StsdBox) (*Codec, error) {
	switch {
	case stsd.AvcX != nil:
		return getAVCCodec(stsd)

	case stsd.HvcX != nil:
		// need example
		spew.Dump(stsd.HvcX)
		return nil, fmt.Errorf("hevc is not yet supported, type: %s", stsd.HvcX.Type())

	case stsd.Mp4a != nil:
		return getMP4ACodec(stsd)

	case stsd.AC3 != nil:
		// need example
		spew.Dump(stsd.AC3)
		return nil, fmt.Errorf("AC3 is not yet supported, type: %s", stsd.AC3.Type())

	case stsd.EC3 != nil:
		// need example
		spew.Dump(stsd.EC3)
		return nil, fmt.Errorf("EC3 is not yet supported, type: %s", stsd.EC3.Type())
	}
	return nil, fmt.Errorf("could not find proper av box in stsd")
}

func getAVCCodec(stsd *mp4ff.StsdBox) (*Codec, error) {
	codecType := stsd.AvcX.Type()
	switch codecType {
	case "avc1", "avc2", "avc3", "avc4":
	default:
		return nil, fmt.Errorf("unknown AvcX codec: %s", codecType)
	}

	// DASH profile string is defined with following formula
	// CODECSTRING = AVCVERSION "." PROFILE CONSTRAINTS LEVEL
	// AVCVERSION = "a" "v" "c" ("1" / "2" / "3" / "4")
	// PROFILE = HEXBYTE
	// CONSTRAINTS = HEXBYTE
	// LEVEL = HEXBYTE
	// https://dashif.org/docs/IOP-Guidelines/DASH-IF-IOP-Part8-v5.0.0.pdf
	//nolint:lll  // https://dvb.org/wp-content/uploads/2020/02/A168r3_MPEG-DASH-Profile-for-Transport-of-ISO-BMFF-Based-DVB-Services_ts_103-285-v140_June_2021.pdf
	// RFC6381 3.3
	return &Codec{
		Profile: fmt.Sprintf("%s.%02x%02x%02x",
			codecType,
			stsd.AvcX.AvcC.AVCProfileIndication,
			stsd.AvcX.AvcC.ProfileCompatibility,
			stsd.AvcX.AvcC.AVCLevelIndication),
	}, nil
}

func getMP4ACodec(stsd *mp4ff.StsdBox) (*Codec, error) {
	codecType := stsd.Mp4a.Type()
	if codecType != "mp4a" {
		return nil, fmt.Errorf("unknown Mp4a codec: %s", codecType)
	}

	// https://dashif.org/docs/IOP-Guidelines/DASH-IF-IOP-Part8-v5.0.0.pdf
	//nolint:lll  // https://dvb.org/wp-content/uploads/2020/02/A168r3_MPEG-DASH-Profile-for-Transport-of-ISO-BMFF-Based-DVB-Services_ts_103-285-v140_June_2021.pdf
	// MPEG-4 AAC Profile audio/mp4 mp4a.40.2 ISO/IEC 14496-14 [9] 1
	// MPEG-4 HE-AAC Profile audio/mp4 mp4a.40.5 ISO/IEC 14496-14 [9] 1
	// MPEG-4 HE-AAC v2 Profile audio/mp4 mp4a.40.29 ISO/IEC 14496-14 [9] 1
	asc, err := aac.DecodeAudioSpecificConfig(bytes.NewReader(
		stsd.Mp4a.Esds.DecConfigDescriptor.DecSpecificInfo.DecConfig))
	if err != nil {
		return nil, fmt.Errorf("cannot decode mp4a AudioSpecificConfig: %w", err)
	}
	var profile string
	switch asc.ObjectType {
	case aac.AAClc:
		profile = "mp4a.40.2"
	case aac.HEAACv1:
		profile = "mp4a.40.5"
	case aac.HEAACv2:
		profile = "mp4a.40.29"
	default:
		return nil, fmt.Errorf("unknown mp4a AudioSpecificConfig ObjectType: %w", err)
	}
	return &Codec{
		Profile:    profile,
		SampleRate: stsd.Mp4a.SampleRate,
	}, nil
}
