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
	saltLen = 10

	bcryptCost = 10

	constrainUID      = "users_uid_key"
	constrainUsername = "users_name_key"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Store struct {
	*store.Database
	salt []byte
}

type Config struct {
	Logger *zap.Logger
	DSN    string
	Salt   string
}

func New(ctx context.Context, cfg *Config) (*Store, error) {
	if len(cfg.Salt) != saltLen {
		return nil, fmt.Errorf("salt length must be %d", saltLen)
	}
	s, err := store.New(ctx, &store.Config{
		Logger:        cfg.Logger,
		DSN:           cfg.DSN,
		Migrate:       true,
		MigrationsDir: &migrations,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create database store: %w", err)
	}
	return &Store{
		Database: s,
		salt:     []byte(cfg.Salt),
	}, nil
}

func (s *Store) Get(ctx context.Context, u *model.User) error {
	hash, err := s.hashPwd(u.Password)
	if err != nil {
		return err
	}
	query := `select uid from users where name = $1 and hash = $2`
	if err = s.Pool().QueryRow(ctx, query, u.Name, hash).Scan(&u.UID); err != nil {
		return handleDBErr(err)
	}
	return nil
}

func (s *Store) Create(ctx context.Context, u *model.User) error {
	hash, err := s.hashPwd(u.Password)
	if err != nil {
		return err
	}
	query := `insert into users (uid, name, hash) values ($1, $2, $3)`
	tag, errDB := s.Pool().Exec(ctx, query, u.UID, u.Name, hash)
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

func (s *Store) salted(pwd string) []byte {
	return append(s.salt, []byte(pwd)...)
}

func (s *Store) hashPwd(pwd string) ([]byte, error) {
	b, err := bcrypt.GenerateFromPassword(s.salted(pwd), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}
	return b, nil
}
