package store

import (
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func (db *Database) migrate(cfg *Config) error {
	if !cfg.Migrate {
		return nil
	}
	if cfg.MigrationsDir == nil {
		return errors.New("migrations dir is not defined")
	}
	db.log.Debug("starting migration")
	change, err := runMigrations(cfg.DSN, cfg.MigrationsDir)
	if err != nil {
		return err
	}
	if change {
		db.log.Info("migration is complete")
	} else {
		db.log.Debug("db is up to date")
	}
	return nil
}

func runMigrations(dsn string, migrationsDir *embed.FS) (bool, error) {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return false, fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return false, fmt.Errorf("failed to get a new migrate instance: %w", err)
	}

	if err = m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return false, fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
		return false, nil
	}
	return true, nil
}
