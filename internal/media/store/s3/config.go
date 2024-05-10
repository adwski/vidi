package s3

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

const (
	defaultRegion = "us-east-1"
)

type StoreConfig struct {
	Logger       *zap.Logger
	Endpoint     string
	AccessKey    string
	SecretKey    string
	Bucket       string
	SSL          bool
	CreateBucket bool
}

func NewStore(cfg *StoreConfig) (*Store, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.SSL,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot init s3 client: %w", err)
	}

	s := &Store{
		client: client,
		logger: cfg.Logger,
		bucket: cfg.Bucket,
	}
	if cfg.CreateBucket {
		if err = s.createBucketIfNotExists(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) createBucketIfNotExists() error {
	ctx := context.Background()
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("cannot check bucket existence: %w", err)
	}
	if exists {
		return nil
	}
	if err = s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{Region: defaultRegion}); err != nil {
		return fmt.Errorf("cannot create bucket: %w", err)
	}
	s.logger.Info("bucket created", zap.String("name", s.bucket))
	return nil
}
