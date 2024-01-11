package main

import (
	"os"

	"github.com/adwski/vidi/internal/app/uploader"
)

func main() {
	os.Exit(uploader.NewApp().Run())
}
