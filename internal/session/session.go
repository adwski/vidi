package session

const (
	KindUpload = "upload"
	KindWatch  = "watch"
)

// Session represents session created for user interactions with media.
type Session struct {
	ID       string `json:"sid"`
	VideoID  string `json:"vid"`
	Location string `json:"loc"`
	PartSize uint64 `json:"psz"`
}
