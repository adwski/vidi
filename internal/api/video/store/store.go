package store

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/adwski/vidi/internal/api/store"
	"github.com/adwski/vidi/internal/api/video/model"
	"github.com/adwski/vidi/internal/mp4/meta"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
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

func (s *Store) DeleteUploadedParts(ctx context.Context, vid string) error {
	query := `delete from upload_parts where video_id = $1`
	_, err := s.Pool().Exec(ctx, query, vid)
	if err != nil {
		return handleDBErr(err)
	}
	return nil
}

func (s *Store) UpdatePart(ctx context.Context, vid string, part *model.Part) error {
	// TODO: may be make all this single pg transaction?
	// This query actually compares base64 encoded checksum strings, but I guess this is ok
	query := `update upload_parts set status = $1 where video_id = $2 and num = $3 and checksum = $4`
	tag, err := s.Pool().Exec(ctx, query, model.PartStatusOK, vid, part.Num, part.Checksum)
	if err != nil {
		return handleDBErr(err)
	}
	if tag.RowsAffected() == 0 {
		// checksum was not ok
		// change status to invalid
		query = `update upload_parts set status = $1 where video_id = $2 and num = $3`
		tag, err = s.Pool().Exec(ctx, query, model.PartStatusInvalid, vid, part.Num)
		if err != nil {
			return handleDBErr(err)
		}
		return nil
	}

	// check if all parts are ok
	// TODO: this query runs on every part update, should check upload completion in some other (more optimal) way
	var notOkCnt uint
	query = `select count(*) as cnt from upload_parts where video_id = $1 and status != $2`
	if err = s.Pool().QueryRow(ctx, query, vid, model.PartStatusOK).Scan(&notOkCnt); err != nil {
		return handleDBErr(err)
	}

	if notOkCnt != 0 {
		// some parts are not ok
		return nil
	}

	// all parts are ok, update video status
	query = `update videos set status = $2 where id = $1`
	tag, err = s.Pool().Exec(ctx, query, vid, int(model.StatusUploaded))
	return handleTagOneRowAndErr(&tag, err)
}

func (s *Store) Usage(ctx context.Context, userID string) (*model.UserUsage, error) {
	usage := &model.UserUsage{}
	query := `select count(*) as v_count, coalesce(sum(size), 0) as v_size from videos where user_id = $1`
	if err := s.Pool().QueryRow(ctx, query, userID).Scan(&usage.Videos, &usage.Size); err != nil {
		return nil, handleDBErr(err)
	}
	return usage, nil
}

func (s *Store) Create(ctx context.Context, vi *model.Video) error {
	batch := &pgx.Batch{}
	batch.Queue(`insert into videos (id, user_id, status, created_at, name, size, location)
		values ($1, $2, $3, $4, $5, $6, $7)`, vi.ID, vi.UserID, int(vi.Status), vi.CreatedAt, vi.Name, vi.Size, vi.Location)
	for _, p := range vi.UploadInfo.Parts {
		batch.Queue(`insert into upload_parts (num, video_id, checksum, status, size)
			values($1, $2, $3, $4, $5)`, p.Num, vi.ID, p.Checksum, p.Status, p.Size)
	}

	if err := s.Pool().SendBatch(ctx, batch).Close(); err != nil {
		return handleDBErr(err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, id, userID string) (*model.Video, error) {
	vi := &model.Video{ID: id, UserID: userID, PlaybackMeta: &meta.Meta{}}
	query := `select location, status, playback_meta, created_at from videos where id = $1 and user_id = $2`
	if err := s.Pool().QueryRow(ctx, query, id, userID).
		Scan(&vi.Location, &vi.Status, &vi.PlaybackMeta, &vi.CreatedAt); err != nil {
		return nil, handleDBErr(err)
	}

	query = `select num, status, size, checksum from upload_parts where video_id = $1`
	rows, err := s.Pool().Query(ctx, query, id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, handleDBErr(err)
		}
		// no upload parts info
		return vi, nil
	}
	vi.UploadInfo = &model.UploadInfo{}
	vi.UploadInfo.Parts, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (*model.Part, error) {
		var part model.Part
		if err = row.Scan(&part.Num, &part.Status, &part.Size, &part.Checksum); err != nil {
			return &part, fmt.Errorf("error while scanning row: %w", err)
		}
		return &part, nil
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, handleDBErr(err)
		}
	}
	return vi, nil
}

func (s *Store) GetAll(ctx context.Context, userID string) ([]*model.Video, error) {
	query := `select id, location, status, name, size, created_at from videos where user_id = $1`
	rows, err := s.Pool().Query(ctx, query, userID)
	if err != nil {
		return nil, handleDBErr(err)
	}
	videos, errR := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*model.Video, error) {
		var vi model.Video
		vi.UserID = userID
		if errS := row.Scan(&vi.ID, &vi.Location, &vi.Status, &vi.Name, &vi.Size, &vi.CreatedAt); errS != nil {
			return nil, fmt.Errorf("error while scanning row: %w", errS)
		}
		return &vi, nil
	})
	if errR != nil {
		return nil, fmt.Errorf("error while collecting rows: %w", errR)
	}
	return videos, nil
}

func (s *Store) Delete(ctx context.Context, id, userID string) error {
	query := `delete from videos where id = $1 and user_id = $2`
	tag, err := s.Pool().Exec(ctx, query, id, userID)
	return handleTagOneRowAndErr(&tag, err)
}

func (s *Store) GetListByStatus(ctx context.Context, status model.Status) ([]*model.Video, error) {
	query := `select id, user_id, location, size, created_at from videos where status = $1`
	rows, err := s.Pool().Query(ctx, query, int(status))
	if err != nil {
		err = model.ErrNotFound
		return nil, err
	}
	videos, errR := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*model.Video, error) {
		var vi model.Video
		vi.Status = status
		if errS := row.Scan(&vi.ID, &vi.UserID, &vi.Location, &vi.Size, &vi.CreatedAt); errS != nil {
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

	for _, vi := range videos {
		query = `select num, status, size, checksum from upload_parts where video_id = $1`
		rows, err = s.Pool().Query(ctx, query, vi.ID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, handleDBErr(err)
			}
		}
		vi.UploadInfo = &model.UploadInfo{}
		vi.UploadInfo.Parts, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (*model.Part, error) {
			var part model.Part
			if err = row.Scan(&part.Num, &part.Status, &part.Size, &part.Checksum); err != nil {
				return &part, fmt.Errorf("error while scanning row: %w", err)
			}
			return &part, nil
		})
		if err != nil {
			return nil, fmt.Errorf("error while collecting rows: %w", errR)
		}
	}
	return videos, nil
}

func (s *Store) UpdateStatus(ctx context.Context, vi *model.Video) error {
	query := `update videos set status = $2 where id = $1`
	tag, err := s.Pool().Exec(ctx, query, vi.ID, int(vi.Status))
	return handleTagOneRowAndErr(&tag, err)
}

func (s *Store) Update(ctx context.Context, vi *model.Video) error {
	query := `update videos set status = $2, playback_meta = $3 where id = $1`
	tag, err := s.Pool().Exec(ctx, query, vi.ID, int(vi.Status), vi.PlaybackMeta)
	return handleTagOneRowAndErr(&tag, err)
}

func handleTagOneRowAndErr(tag *pgconn.CommandTag, err error) error {
	if err != nil {
		return handleDBErr(err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("affected rows: %d, expected: 1", tag.RowsAffected())
	}
	return nil
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
		return model.ErrAlreadyExists
	}
	return fmt.Errorf("postgress error: %w", pgErr)
}
