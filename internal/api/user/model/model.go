package model

import "errors"

var (
	ErrNotFound      = errors.New("user not found")
	ErrAlreadyExists = errors.New("user with this name already exists")
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password,omitempty"`
}

type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewUserFromRequest(id string, req *UserRequest) *User {
	return &User{
		ID:       id,
		Name:     req.Username,
		Password: req.Password,
	}
}
