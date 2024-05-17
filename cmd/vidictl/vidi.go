package main

import (
	"os"

	"github.com/adwski/vidi/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
