package main

import (
	"github.com/adwski/vidi/internal/app/vidictl"
	"os"
)

func main() {
	os.Exit(vidictl.Execute())
}
