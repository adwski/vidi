package model

import (
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("video not found")
	ErrAlreadyExists = errors.New("video with this id already exists")
)

type Video struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Location  string    `json:"loc"`
	Status    Status    `json:"status"`
}

type VideoUpdateRequest struct {
	Status   string `json:"status"`
	Location string `json:"location"`
}

type VideoResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UploadURL string `json:"upload_url,omitempty"`
}

type ListRequest struct {
	Status Status `json:"status"`
}

type WatchResponse struct {
	WatchURL string `json:"watch_url"`
}

func NewVideo(id, userID string) *Video {
	return &Video{
		CreatedAt: time.Now(),
		ID:        id,
		UserID:    userID,
		Status:    VideoStatusCreated,
	}
}

func (v *Video) Response() *VideoResponse {
	return &VideoResponse{
		ID:        v.ID,
		Status:    v.Status.String(),
		CreatedAt: v.CreatedAt.String(),
	}
}

func (v *Video) UploadResponse(url string) *VideoResponse {
	return &VideoResponse{
		ID:        v.ID,
		Status:    v.Status.String(),
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
