// Package event defines generic event that
// sent by processor and uploader to notify videoapi
// about video object changes.
package event

const (
	KindUpdateStatus = iota + 1
	KindVideoPartUploaded
	KindVideoReady
)

// Event is a Video API notification event.
type Event struct {
	PartInfo  *PartInfo
	VideoInfo *VideoInfo
	Kind      int
}

type PartInfo struct {
	VideoID  string
	Checksum string
	Num      uint
}

type VideoInfo struct {
	VideoID  string
	Location string
	Meta     []byte
	Status   int
}
