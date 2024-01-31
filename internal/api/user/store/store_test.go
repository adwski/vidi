package store

import (
	"errors"
	"testing"

	"github.com/adwski/vidi/internal/api/user/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_handleDBErrNotFound(t *testing.T) {
	assert.Equal(t, model.ErrNotFound, handleDBErr(pgx.ErrNoRows))
}

func Test_handleDBErrUnknown(t *testing.T) {
	assert.Contains(t, handleDBErr(errors.New("err")).Error(), "unknown database error")
}

func Test_handleDBErrUnknownPG(t *testing.T) {
	err := &pgconn.PgError{
		Code: "somecode",
	}
	assert.Contains(t, handleDBErr(err).Error(), "postgress error")
}

func Test_handleDBErrAlreadyExists(t *testing.T) {
	err := &pgconn.PgError{
		Code:           pgerrcode.UniqueViolation,
		ConstraintName: constrainUsername,
	}
	assert.Equal(t, model.ErrAlreadyExists, handleDBErr(err))
}

func Test_handleDBErrUIDAlreadyExists(t *testing.T) {
	err := &pgconn.PgError{
		Code:           pgerrcode.UniqueViolation,
		ConstraintName: constrainUID,
	}
	assert.Equal(t, model.ErrUIDAlreadyExists, handleDBErr(err))
}

func TestHashPwdPasswordTooLong(t *testing.T) {
	_, err := hashPwd("qweqweqweqqweqweqweqqweqweqweqqweqweqweqqweqweqweqqweqweqweqqweqweqweq123123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot hash password")
}

func TestHashPwd(t *testing.T) {
	hash, err := hashPwd("qweqwe")
	require.NoError(t, err)
	require.NotEmpty(t, hash)
}

func TestComparePwdErrHash(t *testing.T) {
	err := comparePwd("qwr", "qwe")
	assert.Contains(t, err.Error(), "cannot compare user hash")
}

func TestComparePwdIncorrectCreds(t *testing.T) {
	hash, err := hashPwd("qweqwe")
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	err = comparePwd(hash, "qweqwe1")
	assert.Equal(t, model.ErrIncorrectCredentials, err)
}

func TestComparePwd(t *testing.T) {
	hash, err := hashPwd("qweqwe")
	require.NoError(t, err)
	require.NotEmpty(t, hash)

	err = comparePwd(hash, "qweqwe")
	require.NoError(t, err)
}
