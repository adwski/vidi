package main

import (
	"os"

	"github.com/adwski/vidi/internal/app/video"
)

func main() {
	os.Exit(video.NewApp().Run())
}
