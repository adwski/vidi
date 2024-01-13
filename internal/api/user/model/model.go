package model

import "errors"

var (
	ErrNotFound             = errors.New("user not found")
	ErrAlreadyExists        = errors.New("user with this name already exists")
	ErrUIDAlreadyExists     = errors.New("user with this uid already exists")
	ErrIncorrectCredentials = errors.New("incorrect credentials")
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Password string `json:"password,omitempty"`
}

type UserRequest struct {
	Username string `json:"username" validate:"required,min=4,max=50"`
	Password string `json:"password" validate:"required,min=8,max=50"`
}

func NewUserFromRequest(uid string, req *UserRequest) *User {
	return &User{
		ID:       uid,
		Name:     req.Username,
		Password: req.Password,
	}
}
