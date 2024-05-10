package model

import (
	"errors"
	"github.com/adwski/vidi/internal/mp4/meta"
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

	ErrStaticMPD           = errors.New("unable to generate static mpd")
	ErrInvalidPlaybackMeta = errors.New("invalid playback meta")
	ErrEmptyPlaybackMeta   = errors.New("empty playback meta")
)

type Video struct {
	CreatedAt    time.Time   `json:"created_at"`
	ID           string      `json:"id"`
	UserID       string      `json:"user_id"`
	Name         string      `json:"name"`
	Location     string      `json:"location,omitempty"`
	Status       Status      `json:"status,omitempty"`
	Size         uint64      `json:"size,omitempty"`
	UploadInfo   *UploadInfo `json:"upload_info,omitempty"`
	PlaybackMeta *meta.Meta  `json:"-"`
}

type UploadInfo struct {
	URL   string  `json:"url"`
	Parts []*Part `json:"parts"`
}

type UserStats struct {
	SizeQuota   uint64 `json:"size_total"`
	SizeUsage   uint64 `json:"size_usage"`
	VideosQuota uint   `json:"videos_quota"`
	VideosUsage uint   `json:"videos_usage"`
}

type UserUsage struct {
	Videos uint
	Size   uint64
}

type CreateRequest struct {
	Name  string  `json:"name"`
	Size  uint64  `json:"size_total"`
	Parts []*Part `json:"parts"`
}

func NewVideoNoID(userID, name string, size uint64) *Video {
	return &Video{
		CreatedAt: time.Now().In(time.UTC),
		UserID:    userID,
		Name:      name,
		Size:      size,
		Status:    StatusCreated,
	}
}

func (v *Video) IsReady() bool {
	return v.Status == StatusReady
}

func (v *Video) IsErrored() bool {
	return v.Status == StatusError
}
