package model

import (
	"errors"
	"time"

	"github.com/adwski/vidi/internal/mp4/meta"
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

	ErrNotResumable = errors.New("upload is not resumable")

	ErrInvalidPlaybackMeta = errors.New("invalid playback meta")
	ErrEmptyPlaybackMeta   = errors.New("empty playback meta")
)

type Video struct {
	UploadInfo   *UploadInfo `json:"upload_info,omitempty"`
	PlaybackMeta *meta.Meta  `json:"-"`

	CreatedAt time.Time `json:"created_at"`

	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Location string `json:"location,omitempty"`

	Status Status `json:"status,omitempty"`
	Size   uint64 `json:"size,omitempty"`
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
	Parts []*Part `json:"parts"`
	Size  uint64  `json:"size_total"`
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

func (v *Video) Resumable() bool {
	return v.Status == StatusCreated || v.Status == StatusUploading
}
