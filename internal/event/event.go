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
