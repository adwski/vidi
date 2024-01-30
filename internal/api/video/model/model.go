package model

import (
	"errors"
	"fmt"
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
	Location  string    `json:"loc,omitempty"`
	Status    Status    `json:"status,omitempty"`
}

type UpdateRequest struct {
	Status   string `json:"status"`
	Location string `json:"location"`
}

type Response struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UploadURL string `json:"upload_url,omitempty"`
}

func (r *Response) GetStatus() (Status, error) {
	return GetStatusFromName(r.Status)
}

type ListRequest struct {
	Status Status `json:"status"`
}

func (lr *ListRequest) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"status": "%s"}`, lr.Status.String())), nil
}

type WatchResponse struct {
	WatchURL string `json:"watch_url"`
}

func NewVideo(id, userID string) *Video {
	return &Video{
		CreatedAt: time.Now().In(time.UTC),
		ID:        id,
		UserID:    userID,
		Status:    StatusCreated,
	}
}

func (v *Video) Response() *Response {
	return &Response{
		ID:        v.ID,
		Status:    v.Status.String(),
		CreatedAt: v.CreatedAt.Format(time.RFC3339),
	}
}

func (v *Video) UploadResponse(url string) *Response {
	resp := v.Response()
	resp.UploadURL = url
	return resp
}

func (v *Video) IsReady() bool {
	return v.Status == StatusReady
}

func (v *Video) IsErrored() bool {
	return v.Status == StatusError
}
