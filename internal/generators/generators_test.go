package generators

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/assert"
)

func TestID_Get(t *testing.T) {
	gen := NewID()

	uuid.NewGenWithOptions()

	u, err := gen.Get()
	assert.NoError(t, err)

	b, err := base64.RawURLEncoding.DecodeString(u)
	assert.NoError(t, err)

	_, err = uuid.FromBytes(b)
	assert.NoError(t, err)
}

func TestID_GetError(t *testing.T) {
	gen := NewID()
	gen.gen = uuid.NewGenWithOptions(uuid.WithRandomReader(&errReader{}))

	_, err := gen.Get()
	assert.Error(t, err)
}

func TestID_GetString(t *testing.T) {
	gen := NewID()

	u, err := gen.GetString()
	assert.NoError(t, err)

	_, err = uuid.FromString(u)
	assert.NoError(t, err)
}

func TestID_GetStringError(t *testing.T) {
	gen := NewID()
	gen.gen = uuid.NewGenWithOptions(uuid.WithRandomReader(&errReader{}))

	_, err := gen.GetString()
	assert.Error(t, err)
}

func TestID_GetStringOrPanic(t *testing.T) {
	gen := NewID()

	u := gen.GetStringOrPanic()

	_, err := uuid.FromString(u)
	assert.NoError(t, err)
}

func TestID_GetStringOrPanicPanic(t *testing.T) {
	gen := NewID()
	gen.gen = uuid.NewGenWithOptions(uuid.WithRandomReader(&errReader{}))

	pFunc := func() {
		_ = gen.GetStringOrPanic()
	}

	assert.Panics(t, pFunc)
}

type errReader struct{}

func (e errReader) Read(_ []byte) (int, error) {
	return 0, errors.New("error")
}
