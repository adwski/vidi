package model

import (
	"errors"
	"time"
)

var (
	ErrStorage        = errors.New("storage error")
	ErrInternal       = errors.New("internal error")
	ErrSessionStorage = errors.New("session storage error")
	ErrGivenUp        = errors.New("given up creating video")

	ErrState         = errors.New("video is in error state")
	ErrNotReady      = errors.New("video is not ready")
	ErrNotFound      = errors.New("video not found")
	ErrAlreadyExists = errors.New("video with this id already exists")

	ErrEmptyLocation = errors.New("empty location")
)

type Video struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Location  string    `json:"loc,omitempty"`
	Status    Status    `json:"status,omitempty"`
	UploadURL string    `json:"upload_url,omitempty"`
}

type UserStats struct {
	VideosQuota int   `json:"videos_quota"`
	VideosUsage int   `json:"videos_usage"`
	SizeQuota   int64 `json:"size_total"`
	SizeUsage   int64 `json:"size_usage"`
}

type UserUsage struct {
	Videos int
	Size   int64
}

func NewVideoNoID(userID string) *Video {
	return &Video{
		CreatedAt: time.Now().In(time.UTC),
		UserID:    userID,
		Status:    StatusCreated,
	}
}

func (v *Video) IsReady() bool {
	return v.Status == StatusReady
}

func (v *Video) IsErrored() bool {
	return v.Status == StatusError
}
