package event

import "github.com/adwski/vidi/internal/api/video/model"

const (
	KindUpdateStatus = iota + 1
	KindUpdateLocation
	KindUpdateStatusAndLocation
)

// Event is a Video API notification event.
type Event struct {
	Video model.Video
	Kind  int
}
