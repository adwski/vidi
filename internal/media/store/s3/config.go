package s3

import (
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

type StoreConfig struct {
	Logger    *zap.Logger
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	SSL       bool
}

func NewStore(cfg *StoreConfig) (*Store, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.SSL,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot init s3 client: %w", err)
	}
	return &Store{
		client: client,
		logger: cfg.Logger,
		bucket: cfg.Bucket,
	}, nil
}
