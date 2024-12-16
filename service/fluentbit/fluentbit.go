package fluentbit

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
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

		// Check if the error is specifically because the service is not found
		if err != nil && strings.Contains(string(output), "could not find service") {
			return "[NOT INSTALLED]", nil
		}
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

	case "darwin":
		cmd := exec.Command("sudo", "launchctl", "load", "/Library/LaunchDaemons/fluent-bit.plist")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("launchctl load failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Fluent Bit started successfully.")

	case "windows":
		cmd := exec.Command("sc", "start", "fluent-bit")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("sc start failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Fluent Bit started successfully.")

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return nil
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

	case "darwin":
		cmd := exec.Command("sudo", "launchctl", "unload", "/Library/LaunchDaemons/fluent-bit.plist")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("launchctl unload failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Fluent Bit stopped successfully.")

	case "windows":
		cmd := exec.Command("sc", "stop", "fluent-bit")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("sc stop failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Fluent Bit stopped successfully.")

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return nil
}

// isFluentBitStopped checks if Fluent Bit is no longer running
func isFluentBitStopped() (bool, error) {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("systemctl", "is-active", "fluent-bit")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// If the service is inactive, systemctl returns a non-zero exit code
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() == 3 { // inactive
					return true, nil
				}
			}
			return false, fmt.Errorf("error checking status: %v\nOutput: %s", err, string(output))
		}
		status := string(output)
		return status == "inactive\n", nil

	case "darwin":
		cmd := exec.Command("launchctl", "list", "fluent-bit")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// If the service is not loaded, launchctl returns a non-zero exit code
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() != 0 {
					return true, nil
				}
			}
			return false, fmt.Errorf("error checking status: %v\nOutput: %s", err, string(output))
		}
		// If the service is listed, it's still running
		return false, nil

	case "windows":
		cmd := exec.Command("sc", "query", "fluent-bit")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return false, fmt.Errorf("error checking status: %v\nOutput: %s", err, string(output))
		}
		// Parse the output to check the state
		outputStr := string(output)
		if strings.Contains(outputStr, "STATE              : 1  STOPPED") {
			return true, nil
		}
		return false, nil

	default:
		return false, fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// FluentBitRestart stops and starts Fluent Bit, ensuring it fully stops before restarting
func FluentBitRestart() error {
	// Stop Fluent Bit
	err := FluentBitStop()
	if err != nil {
		return fmt.Errorf("failed to stop Fluent Bit: %v", err)
	}

	// Wait for a short duration to ensure Fluent Bit has stopped
	time.Sleep(2 * time.Second) // Adjust the duration as needed

	// Optionally, verify that Fluent Bit has stopped
	stopped, err := isFluentBitStopped()
	if err != nil {
		return fmt.Errorf("error checking Fluent Bit status: %v", err)
	}
	if !stopped {
		return fmt.Errorf("Fluent Bit did not stop as expected")
	}

	// Start Fluent Bit
	err = FluentBitStart()
	if err != nil {
		return fmt.Errorf("failed to start Fluent Bit: %v", err)
	}

	return nil
}
