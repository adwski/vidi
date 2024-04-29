package main

import (
	"github.com/adwski/vidi/internal/tool"
	"log"
	"os"
)

func main() {
	t, err := tool.New()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(t.Run())
}
