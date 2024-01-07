package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type Store struct {
	logger     *zap.Logger
	client     *minio.Client
	bucket     string
	pathPrefix string
}

func (ms *Store) Put(ctx context.Context, name string, r io.Reader, size int64) error {
	fullName := ms.getFullObjectName(name)
	info, err := ms.client.PutObject(
		ctx,
		ms.bucket,
		fullName,
		r,
		size,
		minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("cannot store object in s3: %w", err)
	}
	ms.logger.Debug("object stored",
		zap.String("location", fullName),
		zap.String("bucket", info.Bucket),
		zap.Int64("size", info.Size))
	return nil
}

func (ms *Store) PutBytes(ctx context.Context, name string, data []byte) error {
	return ms.Put(ctx, name, bytes.NewReader(data), int64(len(data)))
}

func (ms *Store) Get(ctx context.Context, name string) ([]byte, error) {
	fullName := ms.getFullObjectName(name)
	obj, err := ms.client.GetObject(ctx, ms.bucket, fullName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve object from s3: %w", err)
	}
	stat, errS := obj.Stat()
	if errS != nil {
		return nil, fmt.Errorf("cannot get object stats: %w", errS)
	}
	b := make([]byte, stat.Size)
	_, errR := obj.Read(b)
	if errR != nil {
		return nil, fmt.Errorf("cannot read object bytes: %w", errR)
	}
	ms.logger.Debug("object retrieved",
		zap.String("location", fullName),
		zap.String("bucket", ms.bucket),
		zap.Int64("size", stat.Size),
		zap.Time("expiration", stat.Expiration))
	return b, nil
}

func (ms *Store) getFullObjectName(name string) string {
	return fmt.Sprintf("%s%s", ms.pathPrefix, name)
}