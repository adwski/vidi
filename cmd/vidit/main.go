package main

import (
	"log"
	"os"

	"github.com/adwski/vidi/internal/tool"
)

func main() {
	t, err := tool.New()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(t.Run())
}
