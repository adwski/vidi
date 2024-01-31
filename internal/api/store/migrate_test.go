//go:build e2e
// +build e2e

package store

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDatabase_migrateNilMigrationDir(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	db := &Database{
		log: logger,
	}
	err = db.migrate(&Config{
		Logger:        logger,
		MigrationsDir: nil,
		DSN:           "postgres://postgres:postgres@localhost:5432/postgres",
		Migrate:       true,
	})
	require.Error(t, err)
	t.Log(err)
}

func TestDatabase_migrateDisabled(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	db := &Database{
		log: logger,
	}
	err = db.migrate(&Config{
		Logger:        logger,
		MigrationsDir: nil,
		DSN:           "postgres://postgres:postgres@localhost:5432/postgres",
		Migrate:       false,
	})
	require.Nil(t, err)
}

func TestDatabase_migrateInvalidMigrationDir(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	db := &Database{
		log: logger,
	}
	err = db.migrate(&Config{
		Logger:        logger,
		MigrationsDir: &embed.FS{},
		DSN:           "postgres://postgres:postgres@localhost:5432/postgres",
		Migrate:       true,
	})
	require.Error(t, err)
	t.Log(err)
}

//go:embed migrations/*.sql
var migrations embed.FS

func TestDatabase_migrateInvalidDSN(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	db := &Database{
		log: logger,
	}
	err = db.migrate(&Config{
		Logger:        logger,
		MigrationsDir: &migrations,
		DSN:           "qwe://postgres:postgres@localhost:5432/postgres",
		Migrate:       true,
	})
	require.Error(t, err)
	t.Log(err)
}

func TestDatabase_migrateEmptyMigrations(t *testing.T) {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	db := &Database{
		log: logger,
	}
	err = db.migrate(&Config{
		Logger:        logger,
		MigrationsDir: &migrations,
		DSN:           "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		Migrate:       true,
	})
	require.Error(t, err)
	t.Log(err)
}
