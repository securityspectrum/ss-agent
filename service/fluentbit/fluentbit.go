package fluentbit

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// FluentBitStatus checks the status of Fluent Bit using platform-specific commands
func FluentBitStatus() (string, error) {
	fmt.Println("Checking Fluent Bit status...")

	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("systemctl", "is-active", "fluent-bit")
		output, err := cmd.CombinedOutput()
		status := strings.TrimSpace(string(output))

		// Handle systemctl's inactive state without treating it as an error
		if err != nil && strings.Contains(status, "inactive") {
			return "[STOPPED]", nil
		} else if err == nil && strings.Contains(status, "active") {
			return "[RUNNING]", nil
		} else if err != nil {
			return "[FAILED]", fmt.Errorf("systemctl status failed: %v\nOutput: %s", err, string(output))
		}

		return "[UNKNOWN]", nil

	case "darwin":
		cmd := exec.Command("launchctl", "print", "system")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "[FAILED]", fmt.Errorf("launchctl print failed: %v\nOutput: %s", err, string(output))
		}

		if strings.Contains(string(output), "io.fluentbit.fluent-bit") {
			return "[RUNNING]", nil
		}
		return "[STOPPED]", nil

	case "windows":
		cmd := exec.Command("sc", "query", "fluent-bit")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "[FAILED]", fmt.Errorf("sc query failed: %v\nOutput: %s", err, string(output))
		}

		if strings.Contains(string(output), "RUNNING") {
			return "[RUNNING]", nil
		}
		return "[STOPPED]", nil

	default:
		return "[UNKNOWN]", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// FluentBitStart starts Fluent Bit using platform-specific commands
func FluentBitStart() error {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("systemctl", "start", "fluent-bit")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("systemctl start failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Fluent Bit started successfully.")
		return nil

	case "darwin":
		cmd := exec.Command("sudo", "launchctl", "load", "/Library/LaunchDaemons/fluent-bit.plist")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("launchctl load failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Fluent Bit started successfully.")
		return nil

	case "windows":
		cmd := exec.Command("sc", "start", "fluent-bit")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("sc start failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Fluent Bit started successfully.")
		return nil

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// FluentBitStop stops Fluent Bit using platform-specific commands
func FluentBitStop() error {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("systemctl", "stop", "fluent-bit")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("systemctl stop failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Fluent Bit stopped successfully.")
		return nil

	case "darwin":
		cmd := exec.Command("sudo", "launchctl", "unload", "/Library/LaunchDaemons/fluent-bit.plist")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("launchctl unload failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Fluent Bit stopped successfully.")
		return nil

	case "windows":
		cmd := exec.Command("sc", "stop", "fluent-bit")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("sc stop failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Fluent Bit stopped successfully.")
		return nil

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func FluentBitRestart() error {
	// Stop and start Fluent Bit
	err := FluentBitStop()
	if err != nil {
		return err
	}
	return FluentBitStart()
}
