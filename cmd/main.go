// At the moment this is just for testing.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/adwski/vidi/internal/media/store/file"

	"github.com/adwski/vidi/internal/media/processor"
	"github.com/adwski/vidi/internal/mp4"
	"go.uber.org/zap"
)

func main() {
	// testFile := "./testfiles/prog_8s.mp4"
	testFile := "./testfiles/test_seq_h264_high.mp4"
	mp4.Dump(testFile)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("cannot init zap logger:", err)
	}

	outpath := "/Users/artemdvskii/go/src/github.com/adwski/vidi/output"
	mediaStore := file.NewStore(outpath)

	/*
		mediaStore, errS := s3.NewStore(&s3.StoreConfig{
			Endpoint:   "localhost:9000",
			AccessKey:  "admin",
			SecretKey:  "password",
			Bucket:     "vidi",
			PathPrefix: "tests/",
			SSL:        false,
			Logger:     logger,
		})
		if errS != nil {
			log.Fatal("cannot init s3 store:", errS)
		}

	*/

	proc := processor.NewProcessor(logger, mediaStore)

	f, _ := os.Open(testFile)
	defer func() { _ = f.Close() }()
	err = proc.Process(context.Background(), f, "davinci/", time.Second)
	if err != nil {
		fmt.Printf("error processing file: %v\n", err)
	}
}
