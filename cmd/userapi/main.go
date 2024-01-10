package main

import (
	"os"

	"github.com/adwski/vidi/internal/api/app/user"
)

func main() {
	os.Exit(user.NewApp().Run())
}
