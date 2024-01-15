package store

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/adwski/vidi/internal/api/store"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

const (
	constrainUID = "videos_uid_key"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Store struct {
	*store.Database
}

type Config struct {
	Logger *zap.Logger
	DSN    string
}

func New(ctx context.Context, cfg *Config) (*Store, error) {
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
	}, nil
}

func (s *Store) Create(ctx context.Context, vi *model.Video) error {
	query := `insert into videos (id, user_id, status, created_at) values ($1, $2, $3, $4)`
	tag, err := s.Pool().Exec(ctx, query, vi.ID, vi.UserID, int(vi.Status), vi.CreatedAt)
	return handleTagOneRowAndErr(&tag, err)
}

func (s *Store) Get(ctx context.Context, id, userID string) (*model.Video, error) {
	vi := &model.Video{ID: id, UserID: userID}
	query := `select location, status, created_at from videos where id = $1 and user_id = $2`
	if err := s.Pool().QueryRow(ctx, query, id, userID).Scan(&vi.Location, &vi.Status, &vi.CreatedAt); err != nil {
		return nil, handleDBErr(err)
	}
	return vi, nil
}

func (s *Store) Delete(ctx context.Context, id, userID string) error {
	query := `delete from videos where id = $1 and user_id = $2`
	tag, err := s.Pool().Exec(ctx, query, id, userID)
	return handleTagOneRowAndErr(&tag, err)
}

func (s *Store) GetListByStatus(ctx context.Context, status model.Status) ([]*model.Video, error) {
	query := `select id, user_id, location, created_at from videos where status = $1`
	rows, err := s.Pool().Query(ctx, query, int(status))
	if err != nil {
		err = model.ErrNotFound
		return nil, err
	}
	videos, errR := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*model.Video, error) {
		var vi model.Video
		vi.Status = status
		if errS := row.Scan(&vi.ID, &vi.UserID, &vi.Location, &vi.CreatedAt); errS != nil {
			return nil, fmt.Errorf("error while scanning row: %w", errS)
		}
		return &vi, nil
	})
	if errR != nil {
		return nil, fmt.Errorf("error while collecting rows: %w", errR)
	}
	if len(videos) == 0 {
		return nil, model.ErrNotFound
	}
	return videos, nil
}

func (s *Store) UpdateLocation(ctx context.Context, vi *model.Video) error {
	query := `update videos set location = $2 where id = $1`
	tag, err := s.Pool().Exec(ctx, query, vi.ID, vi.Location)
	return handleTagOneRowAndErr(&tag, err)
}

func (s *Store) UpdateStatus(ctx context.Context, vi *model.Video) error {
	query := `update videos set status = $2 where id = $1`
	tag, err := s.Pool().Exec(ctx, query, vi.ID, int(vi.Status))
	return handleTagOneRowAndErr(&tag, err)
}

func (s *Store) Update(ctx context.Context, vi *model.Video) error {
	query := `update videos set status = $2, location = $3 where id = $1`
	tag, err := s.Pool().Exec(ctx, query, vi.ID, int(vi.Status), vi.Location)
	return handleTagOneRowAndErr(&tag, err)
}

func handleTagOneRowAndErr(tag *pgconn.CommandTag, err error) error {
	if err == nil {
		if tag.RowsAffected() != 1 {
			return fmt.Errorf("affected rows: %d, expected: 1", tag.RowsAffected())
		}
		return nil
	}
	return handleDBErr(err)
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
		if pgErr.ConstraintName == constrainUID {
			return model.ErrAlreadyExists
		}
	}
	return fmt.Errorf("postgress error: %w", pgErr)
}
