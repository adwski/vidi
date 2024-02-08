package store

import (
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

	d, err := iofs.New(cfg.MigrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}

	if err = m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
		db.log.Debug("db is up to date")
	} else {
		db.log.Info("migration is complete")
	}
	return nil
}
