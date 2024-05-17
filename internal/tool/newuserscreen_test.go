package tool

import (
	"testing"

	"github.com/labstack/gommon/random"
	"github.com/stretchr/testify/assert"
)

func Test_validateUsernameOK(t *testing.T) {
	assert.NoError(t, validateUsername(random.String(minUserNameLen, "qwerty")))
}

func Test_validateUsernameErr(t *testing.T) {
	assert.Error(t, validateUsername(random.String(minUserNameLen-1, "qwerty")))
}

func Test_validatePasswordOK(t *testing.T) {
	assert.NoError(t, validatePassword(random.String(minPasswordLen, "qwerty")))
}

func Test_validatePasswordErr(t *testing.T) {
	assert.Error(t, validatePassword(random.String(minPasswordLen-1, "qwerty")))
}
