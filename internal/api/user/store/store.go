package store

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/adwski/vidi/internal/api/store"
	"github.com/adwski/vidi/internal/api/user/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 10

	constrainUID      = "users_id"
	constrainUsername = "users_name"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Store struct {
	*store.Database
	logger *zap.Logger
}

type Config struct {
	Logger *zap.Logger
	DSN    string
}

func New(ctx context.Context, cfg *Config) (*Store, error) {
	logger := cfg.Logger.With(zap.String("component", "database"))
	s, err := store.New(ctx, &store.Config{
		Logger:        logger,
		DSN:           cfg.DSN,
		Migrate:       true,
		MigrationsDir: &migrations,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create database store: %w", err)
	}
	return &Store{
		Database: s,
		logger:   logger,
	}, nil
}

func (s *Store) Get(ctx context.Context, u *model.User) error {
	var hash string
	s.logger.Debug("checking user",
		zap.String("name", u.Name),
		zap.String("hash", hash))
	query := `select id, hash from users where name = $1`
	if err := s.Pool().QueryRow(ctx, query, u.Name).Scan(&u.ID, &hash); err != nil {
		return handleDBErr(err)
	}
	return s.compare(hash, u.Password)
}

func (s *Store) Create(ctx context.Context, u *model.User) error {
	hash, err := s.hashPwd(u.Password)
	if err != nil {
		return err
	}
	s.logger.Debug("creating user",
		zap.String("name", u.Name),
		zap.String("id", u.ID),
		zap.String("hash", hash))
	query := `insert into users (id, name, hash) values ($1, $2, $3)`
	tag, errDB := s.Pool().Exec(ctx, query, u.ID, u.Name, hash)
	if errDB == nil {
		if tag.RowsAffected() != 1 {
			return fmt.Errorf("affected rows: %d, expected: 1", tag.RowsAffected())
		}
		return nil
	}
	return handleDBErr(errDB)
}

func handleDBErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return fmt.Errorf("unknown database error: %w", err)
	}
	if pgErr.Code == pgerrcode.UniqueViolation {
		if pgErr.ConstraintName == constrainUsername {
			return model.ErrAlreadyExists
		}
		if pgErr.ConstraintName == constrainUID {
			return model.ErrUIDAlreadyExists
		}
	}
	return fmt.Errorf("postgress error: %w", pgErr)
}

func (s *Store) hashPwd(pwd string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("cannot hash password: %w", err)
	}
	return string(b), nil
}

// compare does 'special' bcrypt-comparison of hashes since we cannot compare them directly.
func (s *Store) compare(hash, pwd string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pwd)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return model.ErrIncorrectCredentials
		}
		return fmt.Errorf("cannot compare user hash: %w", err)
	}
	return nil
}
