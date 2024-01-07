package store

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Database is a relational database storage type.
type Database struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

func (db *Database) Pool() *pgxpool.Pool {
	return db.pool
}

func (db *Database) Close() {
	db.log.Debug("closing pgx connection pool")
	db.pool.Close()
	db.log.Debug("pgx connection pool is closed")
}
