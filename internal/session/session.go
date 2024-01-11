package session

const (
	KindUpload = "upload"
	KindWatch  = "watch"
)

type Session struct {
	ID      string `json:"sid"`
	VideoID string `json:"vid"` // seconds
}
