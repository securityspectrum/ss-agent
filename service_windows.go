// service_windows.go

//go:build windows
// +build windows

package main

import (
	"context"
	"log"
	"ss-agent/cmd"

	"golang.org/x/sys/windows/svc"
)

type myService struct{}

func (m *myService) Execute(args []string, req <-chan svc.ChangeRequest, statusChan chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	statusChan <- svc.Status{State: svc.StartPending}
	statusChan <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	// Create a context for the service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run the agent's main functionality in a goroutine
	go func() {
		cmd.Execute(version, ctx, cancel)
	}()

	// Handle service lifecycle events
	for {
		select {
		case c := <-req:
			switch c.Cmd {
			case svc.Interrogate:
				statusChan <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				statusChan <- svc.Status{State: svc.StopPending}
				cancel()
				return false, 0
			default:
				log.Printf("Unexpected control request: %v", c.Cmd)
			}
		case <-ctx.Done():
			// Service is stopping
			statusChan <- svc.Status{State: svc.Stopped}
			return false, 0
		}
	}
}

func runWindowsService() {
	err := svc.Run("ss-agent", &myService{})
	if err != nil {
		log.Fatalf("Failed to start Windows service: %v", err)
	}
}
