// main_windows.go

//go:build windows
// +build windows

package main

import (
	"context"
	"log"
	"ss-agent/cmd"

	"golang.org/x/sys/windows/svc"
)

var version = "1.0.0"

func main() {
	log.SetFlags(log.LstdFlags)

	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("Failed to determine if running as a Windows service: %v", err)
	}

	if isService {
		// Running as a Windows service
		runWindowsService()
		return
	}

	// Running as a console application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd.Execute(version, ctx, cancel)
}
