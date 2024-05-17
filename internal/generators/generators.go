// Package generators provides simple unique string generator
// that is internally based on uuid v4.
package generators

import (
	"encoding/base64"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

// ID is a UUIDv4 generator. UUID is returned as byte string.
type ID struct {
	gen *uuid.Gen
}

func NewID() *ID {
	return &ID{
		gen: uuid.NewGen(),
	}
}

func (u *ID) Get() (string, error) {
	uu, err := u.gen.NewV4()
	if err != nil {
		return "", fmt.Errorf("cannot generate new uuid: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(uu.Bytes()), nil
}

func (u *ID) GetString() (string, error) {
	uu, err := u.gen.NewV4()
	if err != nil {
		return "", fmt.Errorf("cannot generate new uuid: %w", err)
	}
	return uu.String(), nil
}

func (u *ID) GetStringOrPanic() string {
	uu, err := u.gen.NewV4()
	if err != nil {
		panic(err)
	}
	return uu.String()
}
