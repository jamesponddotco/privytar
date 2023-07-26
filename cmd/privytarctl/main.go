package main

import (
	"os"

	"git.sr.ht/~jamesponddotco/privytar/cmd/privytarctl/internal/app"
)

func main() {
	os.Exit(app.Run())
}
