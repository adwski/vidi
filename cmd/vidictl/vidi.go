package main

import (
	"os"

	"github.com/adwski/vidi/internal/app/vidictl"
)

func main() {
	os.Exit(vidictl.Execute())
}
