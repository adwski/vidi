package store

import (
	"errors"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"

	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
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
		ConstraintName: constrainUID,
	}
	assert.Equal(t, model.ErrAlreadyExists, handleDBErr(err))
}

func Test_handleTagOneRowAndErr(t *testing.T) {
	tag := pgconn.NewCommandTag("qwe1")
	err := handleTagOneRowAndErr(&tag, nil)
	require.NoError(t, err)
}

func Test_handleTagOneRowAndErrWrongAffected(t *testing.T) {
	tag := pgconn.NewCommandTag("qwe11")
	err := handleTagOneRowAndErr(&tag, nil)
	assert.Contains(t, err.Error(), "affected rows: ")
}

func Test_handleTagOneRowAndErrCallHandleDBErr(t *testing.T) {
	tag := pgconn.NewCommandTag("qwe1")
	err := handleTagOneRowAndErr(&tag, errors.New("custom err"))
	assert.Contains(t, err.Error(), "custom err")
}
