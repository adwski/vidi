package meta

import "time"

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
