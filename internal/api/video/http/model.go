package http

import (
	"github.com/adwski/vidi/internal/api/video/model"
	"time"
)

type VideoResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UploadURL string `json:"upload_url,omitempty"`
}

func NewVideoResponse(v *model.Video) *VideoResponse {
	return &VideoResponse{
		ID:        v.ID,
		Status:    v.Status.String(),
		CreatedAt: v.CreatedAt.Format(time.RFC3339),
		UploadURL: v.UploadURL,
	}
}

type ListRequest struct {
	Status string `json:"status"`
}

type UpdateRequest struct {
	Status   string `json:"status"`
	Location string `json:"location"`
}

type WatchResponse struct {
	WatchURL string `json:"watch_url"`
}
