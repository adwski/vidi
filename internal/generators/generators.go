package generators

import (
	"encoding/base64"
	"fmt"

	"github.com/gofrs/uuid/v5"
)

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
