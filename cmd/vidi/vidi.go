package main

import (
	"os"

	"github.com/adwski/vidi/internal/app/vidicli"
)

func main() {
	os.Exit(vidicli.Execute())
}
