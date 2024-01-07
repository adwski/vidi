package model

import "errors"

const (
	MinPasswordLen = 8
	MaxPasswordLen = 50
)

var (
	ErrNotFound         = errors.New("user not found")
	ErrAlreadyExists    = errors.New("user with this name already exists")
	ErrUIDAlreadyExists = errors.New("user with this uid already exists")
)

type User struct {
	UID      string `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password,omitempty"`
}

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewUserFromRequest(uid string, req *UserRequest) *User {
	return &User{
		UID:      uid,
		Name:     req.Username,
		Password: req.Password,
	}
}
