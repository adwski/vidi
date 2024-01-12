package main

import (
	"os"

	"github.com/adwski/vidi/internal/app/processor"
)

func main() {
	os.Exit(processor.NewApp().Run())
}
