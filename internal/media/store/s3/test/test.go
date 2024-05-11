package main

import (
	"bytes"
	"context"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	client, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("admin", "password", ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	b := []byte("test")
	buf := bytes.NewBuffer(b)
	ctx := context.Background()

	info, err := client.PutObject(ctx, "vidi", "test/test", buf, int64(len(b)), minio.PutObjectOptions{
		SendContentMd5:       true,
		DisableContentSha256: false,
	})
	spew.Dump(info, err)
}
