package s3

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/sha256-simd"
	"go.uber.org/zap"
)

var ErrNotFount = errors.New("not found")

// Store is media store that uses s3 compatible storage.
// It implements simple Get/Set operations, and in addition
// it can calculate sha256 checksum of already uploaded object.
//
// TODO Seems like S3 API actually can calculate sha256 on server side,
// TODO but I couldn't get it working with minio. Need to investigate further.
type Store struct {
	logger *zap.Logger
	client *minio.Client
	bucket string
}

func (s *Store) Put(ctx context.Context, name string, r io.Reader, size int64) error {
	_, err := s.client.PutObject(ctx, s.bucket, name, r, size, minio.PutObjectOptions{DisableMultipart: true})
	if err != nil {
		return fmt.Errorf("cannot store object in s3: %w", err)
	}
	return nil
}

func (s *Store) CalcSha256(ctx context.Context, name string) (string, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("cannot retrieve object from s3: %w", err)
	}
	defer func() { _ = obj.Close() }()
	shaW := sha256.New()
	_, err = io.Copy(shaW, obj)
	if err != nil {
		return "", fmt.Errorf("cannot calculate object sha256: %w", err)
	}
	return base64.StdEncoding.EncodeToString(shaW.Sum(nil)), nil
}

func (s *Store) Get(ctx context.Context, name string) (io.ReadSeekCloser, int64, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("cannot retrieve object from s3: %w", err)
	}
	stat, errS := obj.Stat()
	if errS != nil {
		er := minio.ToErrorResponse(errS)
		if er.StatusCode == http.StatusNotFound {
			return nil, 0, ErrNotFount
		}
		return nil, 0, fmt.Errorf("cannot get object stats: %w", errS)
	}
	return obj, stat.Size, nil
}
