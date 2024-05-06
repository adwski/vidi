package event

const (
	KindUpdateStatus = iota + 1
	KindUpdateStatusAndLocation
	KindVideoPartUploaded
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
	Status   int
}
