package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type Store struct {
	logger *zap.Logger
	client *minio.Client
	bucket string
}

func (ms *Store) Put(ctx context.Context, name string, r io.Reader, size int64) error {
	info, err := ms.client.PutObject(ctx, ms.bucket, name, r, size, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("cannot store object in s3: %w", err)
	}
	ms.logger.Debug("object stored",
		zap.String("location", name),
		zap.String("bucket", info.Bucket),
		zap.Int64("size", info.Size))
	return nil
}

func (ms *Store) GetBytes(ctx context.Context, name string) ([]byte, error) {
	rc, size, err := ms.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	b := make([]byte, size)
	_, errR := rc.Read(b)
	if errR != nil {
		return nil, fmt.Errorf("cannot read object bytes: %w", errR)
	}
	ms.logger.Debug("object retrieved",
		zap.String("location", name),
		zap.String("bucket", ms.bucket),
		zap.Int64("size", size))
	return b, nil
}

func (ms *Store) Get(ctx context.Context, name string) (io.ReadCloser, int64, error) {
	obj, err := ms.client.GetObject(ctx, ms.bucket, name, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("cannot retrieve object from s3: %w", err)
	}
	stat, errS := obj.Stat()
	if errS != nil {
		return nil, 0, fmt.Errorf("cannot get object stats: %w", errS)
	}
	return obj, stat.Size, nil
}
