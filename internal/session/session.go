package session

type Session struct {
	ID      string `json:"sid"`
	VideoID string `json:"vid"` // seconds
}
