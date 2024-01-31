//go:build e2e
// +build e2e

package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDatabase_UpDown(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	db, err := New(context.Background(), &Config{
		Logger: logger,
		DSN:    "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
	})
	require.NoError(t, err)
	db.Close()
}
