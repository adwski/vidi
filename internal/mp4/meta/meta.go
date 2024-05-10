package meta

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	jsoniter "github.com/json-iterator/go"
	"time"
)

var (
	jEnc = jsoniter.ConfigCompatibleWithStandardLibrary
)

// Meta is a generic media file structure.
type Meta struct {
	Tracks   []Track
	Duration time.Duration
}

// Track is a media file video or audio track.
type Track struct {
	Codec     *Codec
	Segment   *SegmentConfig
	Name      string
	MimeType  string
	Bandwidth uint32
}

// SegmentConfig holds segment related info of segmented media file.
type SegmentConfig struct {
	Init        string
	StartNumber uint
	Duration    uint64
	Timescale   uint32
}

func (mt *Meta) TextValue() (pgtype.Text, error) {
	b, err := jEnc.Marshal(mt)
	if err != nil {
		return pgtype.Text{}, fmt.Errorf("failed to encode meta: %w", err)
	}
	return pgtype.Text{String: string(b), Valid: true}, nil
}

func (mt *Meta) ScanText(t pgtype.Text) error {
	return jEnc.Unmarshal([]byte(t.String), mt)
}
