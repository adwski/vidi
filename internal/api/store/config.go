package store

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	defaultConnectTimeout           = 3 * time.Second
	defaultConnectionIdle           = time.Minute
	defaultConnectionLifeTime       = time.Hour
	defaultConnectionLifeTimeJitter = 5 * time.Minute
	defaultHealthCheckPeriod        = 3 * time.Second
	defaultMaxConns                 = 5
	defaultMinConns                 = 2
)

type Config struct {
	Logger        *zap.Logger
	MigrationsDir *embed.FS
	DSN           string
	Migrate       bool
}

func New(ctx context.Context, cfg *Config) (*Database, error) {
	if cfg.Logger == nil {
		return nil, errors.New("nil logger")
	}
	logger := cfg.Logger.With(zap.String("component", "database"))

	pCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("cannot parse DSN: %w", err)
	}
	preparePoolConfig(pCfg)

	db := &Database{
		log: logger,
	}
	if err = db.init(ctx, cfg, pCfg); err != nil {
		return nil, err
	}
	return db, nil
}

func preparePoolConfig(pCfg *pgxpool.Config) {
	pCfg.ConnConfig.Config.ConnectTimeout = defaultConnectTimeout

	// Choosing this mode because:
	// - Compatible with connection pollers
	// - Does not make two round trips
	// - Does not imply side effects after schema change
	// - We're using simple data types
	pCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	pCfg.MaxConnLifetime = defaultConnectionLifeTime
	pCfg.MaxConnLifetimeJitter = defaultConnectionLifeTimeJitter
	pCfg.MaxConnIdleTime = defaultConnectionIdle
	pCfg.MaxConns = defaultMaxConns
	pCfg.MinConns = defaultMinConns
	pCfg.HealthCheckPeriod = defaultHealthCheckPeriod
}

func (db *Database) init(ctx context.Context, cfg *Config, pCfg *pgxpool.Config) (err error) {
	if err = db.migrate(cfg); err != nil {
		return
	}

	if db.pool, err = pgxpool.NewWithConfig(ctx, pCfg); err != nil {
		err = fmt.Errorf("cannot create pgx connection pool: %w", err)
	}
	return
}
