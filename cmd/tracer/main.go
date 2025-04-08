package main

import (
	"fmt"
	"os"

	"github.com/helmedeiros/tracer-bullet/internal/commands"
)

func main() {
	if err := commands.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
