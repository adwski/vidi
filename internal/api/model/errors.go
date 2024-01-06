package model

const (
	InternalError = "internal error"
)

type Response struct {
	Message string `json:"msg,omitempty"`
	Error   string `json:"error,omitempty"`
}
