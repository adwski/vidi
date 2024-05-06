package s3

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/adwski/vidi/internal/media/store"

	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

// Store is media store that uses s3 compatible storage.
type Store struct {
	logger *zap.Logger
	client *minio.Client
	bucket string
}

func (s *Store) Put(ctx context.Context, name string, r io.Reader, size int64) error {
	info, err := s.client.PutObject(ctx, s.bucket, name, r, size, minio.PutObjectOptions{DisableMultipart: true})
	if err != nil {
		return fmt.Errorf("cannot store object in s3: %w", err)
	}
	s.logger.Debug("object stored",
		zap.String("location", name),
		zap.String("bucket", info.Bucket),
		zap.Int64("size", info.Size))
	return nil
}

func (s *Store) Get(ctx context.Context, name string) (io.ReadCloser, int64, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("cannot retrieve object from s3: %w", err)
	}
	stat, errS := obj.Stat()
	if errS != nil {
		er := minio.ToErrorResponse(errS)
		if er.StatusCode == http.StatusNotFound {
			return nil, 0, store.ErrNotFount
		}
		return nil, 0, fmt.Errorf("cannot get object stats: %w", errS)
	}
	return obj, stat.Size, nil
}

func (s *Store) GetChecksumSHA256(ctx context.Context, name string) (string, error) {
	info, err := s.client.StatObject(ctx, s.bucket, name, minio.StatObjectOptions{})
	if err != nil {
		er := minio.ToErrorResponse(err)
		if er.StatusCode == http.StatusNotFound {
			return "", store.ErrNotFount
		}
	}
	return info.ChecksumSHA256, nil
}
