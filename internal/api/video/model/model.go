package model

import (
	"errors"
	"time"

	user "github.com/adwski/vidi/internal/api/user/model"
)

// Video statuses.
// TODO: May be transitive statuses are redundant?
// TODO: How these statuses will map to DB enums? (or may be don't use enums?)
const (
	VideoStatusError = iota - 1
	VideoStatusCreated
	VideoStatusUploading
	VideoStatusUploaded
	VideoStatusProcessing
	VideoStatusReady
)

var (
	ErrNotFound      = errors.New("video not found")
	ErrAlreadyExists = errors.New("video with this id already exists")
)

type VideoResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UploadURL string `json:"upload_url,omitempty"`
}

type WatchResponse struct {
	WatchURL string `json:"watch_url"`
}

type Video struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Location  string    `json:"loc"`
	Status    int       `json:"status"`
}

func NewVideo(id string, u *user.User) *Video {
	return &Video{
		CreatedAt: time.Now(),
		ID:        id,
		UserID:    u.ID,
		Status:    VideoStatusCreated,
	}
}

func (v *Video) Response() *VideoResponse {
	return &VideoResponse{
		ID:        v.ID,
		Status:    v.StatusName(),
		CreatedAt: v.CreatedAt.String(),
	}
}

func (v *Video) UploadResponse(url string) *VideoResponse {
	return &VideoResponse{
		ID:        v.ID,
		Status:    v.StatusName(),
		CreatedAt: v.CreatedAt.String(),
		UploadURL: url,
	}
}

func (v *Video) IsReady() bool {
	return v.Status == VideoStatusReady
}

func (v *Video) IsErrored() bool {
	return v.Status == VideoStatusError
}

func (v *Video) StatusName() string {
	switch v.Status {
	case VideoStatusError:
		return "error"
	case VideoStatusCreated:
		return "created"
	case VideoStatusUploading:
		return "uploading"
	case VideoStatusUploaded:
		return "uploaded"
	case VideoStatusProcessing:
		return "processing"
	case VideoStatusReady:
		return "ready"
	default:
		return "unknown"
	}
}
