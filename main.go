// main.go
//go:build linux || darwin
// +build linux darwin

package main

import (
	"context"
	"log"

	"ss-agent/cmd"
)

var version = "1.0.0"

func main() {
	// Configure logging to include timestamps
	log.SetFlags(log.LstdFlags)

	// Set up context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Execute the root command with version and context
	cmd.Execute(version, ctx, cancel)
}
