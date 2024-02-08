package main

import (
	"os"

	"github.com/adwski/vidi/internal/app/streamer"
)

func main() {
	os.Exit(streamer.NewApp().Run())
}
