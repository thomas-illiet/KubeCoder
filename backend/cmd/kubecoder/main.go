package main

import (
	"fmt"
	"os"

	"github.com/thomas-illiet/KubeCoder/backend/internal/command"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// main executes the KubeCoder command-line application.
func main() {
	if err := command.New(version, commit, date).Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
