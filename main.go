package main

import (
	"os"

	"github.com/suer/first-task-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
