// service/service.go

package service

import (
	"fmt"
	"os/exec"
	"ss-agent/utils/osinfo"
	"strings"
)

// AllServices contains all the services managed by ss-agent.
var AllServices = []string{"fluent-bit", "zeek", "osquery"}

// InstallService installs the specified service based on the OS type and distribution.
func InstallService(serviceName string) error {
	osType := osinfo.GetOSType()
	osDist := osinfo.GetOSDist()
	fmt.Printf("Installing %s on %s/%s...\n", serviceName, osType, osDist)
	var installCmd string

	switch osType {
	case "windows":
		switch serviceName {
		case "fluent-bit":
			installCmd = "choco install fluent-bit"
		case "zeek":
			installCmd = "choco install zeek"
		case "osquery":
			installCmd = "choco install osquery"
		default:
			return fmt.Errorf("unknown service: %s", serviceName)
		}
	case "darwin":
		switch serviceName {
		case "fluent-bit":
			installCmd = "brew install fluent-bit"
		case "zeek":
			installCmd = "brew install zeek"
		case "osquery":
			installCmd = "brew install osquery"
		default:
			return fmt.Errorf("unknown service: %s", serviceName)
		}
	case "linux":
		switch osDist {
		case "ubuntu", "debian", "mint":
			switch serviceName {
			case "fluent-bit":
				installCmd = "sudo apt-get install -y fluent-bit"
			case "zeek":
				installCmd = "sudo apt-get install -y zeek"
			case "osquery":
				installCmd = "sudo apt-get install -y osquery"
			default:
				return fmt.Errorf("unknown service: %s", serviceName)
			}
		case "fedora", "rhel", "centos":
			switch serviceName {
			case "fluent-bit":
				installCmd = "sudo dnf install -y fluent-bit"
			case "zeek":
				installCmd = "sudo dnf install -y zeek"
			case "osquery":
				installCmd = "sudo dnf install -y osquery"
			default:
				return fmt.Errorf("unknown service: %s", serviceName)
			}
		default:
			return fmt.Errorf("unsupported distribution: %s", osDist)
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", osType)
	}

	cmd := exec.Command("sh", "-c", installCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s: %v\nOutput: %s", serviceName, err, string(output))
	}
	fmt.Printf("%s installed successfully on %s/%s.\n", serviceName, osType, osDist)
	return nil
}

// UninstallService uninstalls the specified service based on the OS type and distribution.
func UninstallService(serviceName string) error {
	osType := osinfo.GetOSType()
	osDist := osinfo.GetOSDist()
	fmt.Printf("Uninstalling %s on %s/%s...\n", serviceName, osType, osDist)
	var uninstallCmd string

	switch osType {
	case "windows":
		switch serviceName {
		case "fluent-bit":
			uninstallCmd = "choco uninstall fluent-bit"
		case "zeek":
			uninstallCmd = "choco uninstall zeek"
		case "osquery":
			uninstallCmd = "choco uninstall osquery"
		default:
			return fmt.Errorf("unknown service: %s", serviceName)
		}
	case "darwin":
		switch serviceName {
		case "fluent-bit":
			uninstallCmd = "brew uninstall fluent-bit"
		case "zeek":
			uninstallCmd = "brew uninstall zeek"
		case "osquery":
			uninstallCmd = "brew uninstall osquery"
		default:
			return fmt.Errorf("unknown service: %s", serviceName)
		}
	case "linux":
		switch osDist {
		case "ubuntu", "debian", "mint":
			switch serviceName {
			case "fluent-bit":
				uninstallCmd = "sudo apt-get remove -y fluent-bit"
			case "zeek":
				uninstallCmd = "sudo apt-get remove -y zeek"
			case "osquery":
				uninstallCmd = "sudo apt-get remove -y osquery"
			default:
				return fmt.Errorf("unknown service: %s", serviceName)
			}
		case "fedora", "rhel", "centos":
			switch serviceName {
			case "fluent-bit":
				uninstallCmd = "sudo dnf remove -y fluent-bit"
			case "zeek":
				uninstallCmd = "sudo dnf remove -y zeek"
			case "osquery":
				uninstallCmd = "sudo dnf remove -y osquery"
			default:
				return fmt.Errorf("unknown service: %s", serviceName)
			}
		default:
			return fmt.Errorf("unsupported distribution: %s", osDist)
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", osType)
	}

	cmd := exec.Command("sh", "-c", uninstallCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall %s: %v\nOutput: %s", serviceName, err, string(output))
	}
	fmt.Printf("%s uninstalled successfully on %s/%s.\n", serviceName, osType, osDist)
	return nil
}

// ManageService manages the specified service based on the action.
// If serviceName is "all", it applies the action to all services.
func ManageService(serviceName, action string) error {
	osType := osinfo.GetOSType()
	osDist := osinfo.GetOSDist()

	if strings.ToLower(serviceName) == "all" {
		// Apply the action to all services
		fmt.Printf("%s all services on %s/%s...\n", strings.Title(action), osType, osDist)
		var overallError error
		for _, svc := range AllServices {
			err := performAction(svc, action, osType, osDist)
			if err != nil {
				fmt.Printf("Error %sing %s: %v\n", action, svc, err)
				overallError = err
			} else {
				fmt.Printf("%s %s service successfully on %s/%s.\n", strings.Title(action), svc, osType, osDist)
			}
		}
		return overallError
	}

	// Continue with individual service management if service name is provided
	fmt.Printf("%s %s service on %s/%s...\n", strings.Title(action), serviceName, osType, osDist)
	err := performAction(serviceName, action, osType, osDist)
	if err != nil {
		return fmt.Errorf("failed to %s %s: %v", action, serviceName, err)
	}
	fmt.Printf("%s %s service successfully on %s/%s.\n", strings.Title(action), serviceName, osType, osDist)
	return nil
}

// performAction performs the specified action on a single service.
func performAction(serviceName, action, osType, osDist string) error {
	var manageCmd string

	switch osType {
	case "windows":
		switch action {
		case "start", "stop", "restart", "status":
			// Windows services often have different naming conventions; adjust if necessary
			serviceMap := map[string]string{
				"osquery":    "osqueryd",
				"fluent-bit": "fluent-bit",
				"zeek":       "zeek",
			}
			mappedService, exists := serviceMap[serviceName]
			if !exists {
				return fmt.Errorf("unknown service: %s", serviceName)
			}
			if action == "restart" {
				manageCmd = fmt.Sprintf("sc stop %s && sc start %s", mappedService, mappedService)
			} else {
				manageCmd = fmt.Sprintf("sc %s %s", action, mappedService)
			}
		default:
			return fmt.Errorf("unsupported action: %s for Windows", action)
		}
	case "darwin":
		switch action {
		case "start", "stop", "restart", "status":
			serviceMap := map[string]string{
				"osquery":    "com.facebook.osqueryd",
				"fluent-bit": "homebrew.mxcl.fluent-bit",
				"zeek":       "org.zeek.zeek",
			}
			mappedService, exists := serviceMap[serviceName]
			if !exists {
				return fmt.Errorf("unknown service: %s", serviceName)
			}
			// Note: macOS uses launchctl with different syntax
			if action == "restart" {
				manageCmd = fmt.Sprintf("launchctl stop %s && launchctl start %s", mappedService, mappedService)
			} else {
				manageCmd = fmt.Sprintf("launchctl %s %s", action, mappedService)
			}
		default:
			return fmt.Errorf("unsupported action: %s for macOS", action)
		}
	case "linux":
		switch action {
		case "start", "stop", "restart", "status":
			serviceMap := map[string]string{
				"osquery":    "osqueryd",
				"fluent-bit": "fluent-bit",
				"zeek":       "zeek",
			}
			mappedService, exists := serviceMap[serviceName]
			if !exists {
				return fmt.Errorf("unknown service: %s", serviceName)
			}
			if action == "restart" {
				manageCmd = fmt.Sprintf("sudo systemctl restart %s", mappedService)
			} else {
				manageCmd = fmt.Sprintf("sudo systemctl %s %s", action, mappedService)
			}
		default:
			return fmt.Errorf("unsupported action: %s for Linux", action)
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", osType)
	}

	cmd := exec.Command("sh", "-c", manageCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s failed: %v\nOutput: %s", action, err, string(output))
	}
	return nil
}

// CheckServiceHealth checks the health of the specified service.
func CheckServiceHealth(serviceName string) (string, error) {
	fmt.Printf("Checking health of %s...\n", serviceName)
	var healthCmd string

	switch serviceName {
	case "fluent-bit":
		healthCmd = "systemctl is-active fluent-bit"
	case "zeek":
		healthCmd = "systemctl is-active zeek"
	case "osquery":
		healthCmd = "systemctl is-active osqueryd"
	default:
		return "", fmt.Errorf("unknown service: %s", serviceName)
	}

	cmd := exec.Command("sh", "-c", healthCmd)
	output, err := cmd.CombinedOutput()
	status := strings.TrimSpace(string(output))
	if err != nil {
		return status, fmt.Errorf("failed to check health of %s: %v\nOutput: %s", serviceName, err, string(output))
	}

	return status, nil
}

// CheckServiceStatus checks the current status of the specified service.
func CheckServiceStatus(serviceName string) (string, error) {
	osType := osinfo.GetOSType()
	osDist := osinfo.GetOSDist()
	fmt.Printf("Checking status of %s on %s/%s...\n", serviceName, osType, osDist)
	var statusCmd string

	switch osType {
	case "windows":
		serviceMap := map[string]string{
			"osquery":    "osqueryd",
			"fluent-bit": "fluent-bit",
			"zeek":       "zeek",
		}
		mappedService, exists := serviceMap[serviceName]
		if !exists {
			return "", fmt.Errorf("unknown service: %s", serviceName)
		}
		statusCmd = fmt.Sprintf("sc query %s | findstr /I \"STATE\"", mappedService)
	case "darwin":
		serviceMap := map[string]string{
			"osquery":    "com.facebook.osqueryd",
			"fluent-bit": "homebrew.mxcl.fluent-bit",
			"zeek":       "org.zeek.zeek",
		}
		mappedService, exists := serviceMap[serviceName]
		if !exists {
			return "", fmt.Errorf("unknown service: %s", serviceName)
		}
		statusCmd = fmt.Sprintf("launchctl list | grep %s", mappedService)
	case "linux":
		serviceMap := map[string]string{
			"osquery":    "osqueryd",
			"fluent-bit": "fluent-bit",
			"zeek":       "zeek",
		}
		mappedService, exists := serviceMap[serviceName]
		if !exists {
			return "", fmt.Errorf("unknown service: %s", serviceName)
		}
		statusCmd = fmt.Sprintf("systemctl is-active %s", mappedService)
	default:
		return "", fmt.Errorf("unsupported operating system: %s", osType)
	}

	cmd := exec.Command("sh", "-c", statusCmd)
	output, err := cmd.CombinedOutput()
	status := strings.TrimSpace(string(output))

	if osType == "windows" {
		// Parse Windows SC query output
		if strings.Contains(status, "RUNNING") {
			status = "active"
		} else if strings.Contains(status, "STOPPED") {
			status = "inactive"
		} else {
			status = "unknown"
		}
	} else if osType == "darwin" {
		if status == "" {
			status = "inactive"
		} else {
			status = "active"
		}
	} else if osType == "linux" {
		// systemctl already returns active/inactive/failed
	}

	if err != nil && status == "" {
		return status, fmt.Errorf("failed to check status of %s: %v\nOutput: %s", serviceName, err, string(output))
	}

	return status, nil
}

// RenderHealthStatus displays the health status of a service in a formatted manner.
func RenderHealthStatus(serviceName, status string) {
	var statusText string
	switch status {
	case "active":
		statusText = "[RUNNING]"
	case "inactive":
		statusText = "[STOPPED]"
	case "failed":
		statusText = "[FAILED]"
	default:
		statusText = "[UNKNOWN]"
	}
	fmt.Printf("%-15s: %s\n", serviceName, statusText)
}

// HealthCheck performs a health check on a specified service or all services if 'all' is provided.
func HealthCheck(serviceName string) {
	if strings.ToLower(serviceName) == "all" {
		// List all services' statuses
		fmt.Printf("Listing all service statuses on %s/%s\n", osinfo.GetOSType(), osinfo.GetOSDist())
		for _, svc := range AllServices {
			status, err := CheckServiceStatus(svc)
			if err != nil {
				fmt.Printf("Error checking status of %s: %v\n", svc, err)
			} else {
				RenderHealthStatus(svc, status)
			}
		}
	} else {
		// Check the status of a specific service
		status, err := CheckServiceStatus(serviceName)
		if err != nil {
			fmt.Printf("Error checking status of %s: %v\n", serviceName, err)
		} else {
			RenderHealthStatus(serviceName, status)
		}
	}
}
