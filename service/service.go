package service

import (
	"fmt"
	"os/exec"
	"ss-agent/utils/osinfo"
	"strings"
)

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

func ManageService(serviceName, action string) error {
	osType := osinfo.GetOSType()
	osDist := osinfo.GetOSDist()
	fmt.Printf("%s %s service on %s/%s...\n", strings.Title(action), serviceName, osType, osDist)
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
			manageCmd = fmt.Sprintf("sc %s %s", action, mappedService)
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
			manageCmd = fmt.Sprintf("sudo systemctl %s %s", action, mappedService)
		default:
			return fmt.Errorf("unsupported action: %s for Linux", action)
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", osType)
	}

	cmd := exec.Command("sh", "-c", manageCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to %s %s: %v\nOutput: %s", action, serviceName, err, string(output))
	}
	fmt.Printf("%s %s service successfully on %s/%s.\n", strings.Title(action), serviceName, osType, osDist)
	return nil
}

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

func HealthCheck() {
	services := []string{"fluent-bit", "zeek", "osquery"}
	for _, service := range services {
		status, err := CheckServiceHealth(service)
		if err != nil {
			RenderHealthStatus(service, status)
		} else {
			RenderHealthStatus(service, status)
		}
	}
}
