// service/service.go

package service

import (
	"fmt"
	"ss-agent/service/fluentbit"
	"ss-agent/service/osquery"
	"ss-agent/service/zeek"
	"strings"
)

var AllServices = []string{"fluent-bit", "zeek", "osquery"}

// ManageService manages the specified service based on the action.
func ManageService(serviceName, action string) error {
	switch strings.ToLower(serviceName) {
	case "zeek":
		return handleZeekService(action)
	case "fluent-bit":
		return handleFluentBitService(action)
	default:
		return fmt.Errorf("unknown service: %s", serviceName)
	}
}

// handleZeekService handles the actions for the Zeek service
func handleZeekService(action string) error {
	switch action {
	case "start":
		return zeek.ZeekStart()
	case "stop":
		return zeek.ZeekStop()
	case "restart":
		return zeek.ZeekRestart() // Restart calls both stop and start
	case "status":
		status, err := zeek.ZeekStatus()
		if err != nil {
			return err
		}
		fmt.Printf("Zeek: %s\n", status)
		return nil
	default:
		return fmt.Errorf("unknown action: %s for Zeek", action)
	}
}

// handleFluentBitService handles the actions for the Fluent Bit service
func handleFluentBitService(action string) error {
	switch action {
	case "start":
		return fluentbit.FluentBitStart()
	case "stop":
		return fluentbit.FluentBitStop()
	case "restart":
		return fluentbit.FluentBitRestart() // Restart calls both stop and start
	case "status":
		status, err := fluentbit.FluentBitStatus()
		if err != nil {
			return err
		}
		fmt.Printf("Fluent Bit: %s\n", status)
		return nil
	default:
		return fmt.Errorf("unknown action: %s for Fluent Bit", action)
	}
}

func HealthCheck(serviceName string) {
	if strings.ToLower(serviceName) == "all" {
		// List all services' statuses
		fmt.Println("Listing all service statuses...")
		for _, svc := range AllServices {
			status, err := checkServiceStatus(svc)
			if err != nil {
				fmt.Printf("%-15s: [ERROR] %v\n", svc, err)
			} else {
				fmt.Printf("%-15s: %s\n", svc, status)
			}
		}
	} else {
		// Check the status of a specific service
		status, err := checkServiceStatus(serviceName)
		if err != nil {
			fmt.Printf("%-15s: [ERROR] %v\n", serviceName, err)
		} else {
			fmt.Printf("%-15s: %s\n", serviceName, status)
		}
	}
}

func checkServiceStatus(serviceName string) (string, error) {
	switch strings.ToLower(serviceName) {
	case "zeek":
		status, err := zeek.ZeekStatus() // Capture both status and error
		if err != nil {
			return "[UNKNOWN]", err // Return error if there is one
		}
		return status, nil // Return status if no error
	case "fluent-bit":
		status, err := fluentbit.FluentBitStatus() // Capture both status and error
		if err != nil {
			return "[UNKNOWN]", err
		}
		return status, nil
	case "osquery":
		status, err := osquery.OsqueryStatus() // Capture both status and error
		if err != nil {
			return "[UNKNOWN]", err
		}
		return status, nil
	default:
		return "[UNKNOWN]", fmt.Errorf("unknown service: %s", serviceName)
	}
}
