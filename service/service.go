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
	fmt.Printf("%s %s service...\n", action, serviceName)
	var manageCmd string

	switch serviceName {
	case "fluent-bit":
		manageCmd = fmt.Sprintf("sudo systemctl %s fluent-bit", action)
	case "zeek":
		manageCmd = fmt.Sprintf("sudo systemctl %s zeek", action)
	case "osquery":
		manageCmd = fmt.Sprintf("sudo systemctl %s osqueryd", action)
	default:
		return fmt.Errorf("unknown service: %s", serviceName)
	}

	cmd := exec.Command("sh", "-c", manageCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to %s %s: %v\nOutput: %s", action, serviceName, err, string(output))
	}
	fmt.Printf("%s %s service successfully.\n", strings.Title(action), serviceName)
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
