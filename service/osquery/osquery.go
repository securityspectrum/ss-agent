package osquery

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// OsqueryStatus checks the status of Osquery using platform-specific commands
func OsqueryStatus() (string, error) {
	fmt.Println("Checking osqueryd status...")

	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("systemctl", "is-active", "osqueryd")
		output, err := cmd.CombinedOutput()
		status := strings.TrimSpace(string(output))

		// If systemctl returns an error, but the output is "inactive", treat it as STOPPED
		if err != nil && strings.Contains(status, "inactive") {
			return "[STOPPED]", nil
		} else if err == nil && strings.Contains(status, "active") {
			return "[RUNNING]", nil
		} else if err != nil {
			return "[FAILED]", fmt.Errorf("systemctl status failed: %v\nOutput: %s", err, string(output))
		}

		return "[UNKNOWN]", nil

	case "darwin":
		cmd := exec.Command("launchctl", "list", "com.facebook.osqueryd")
		output, err := cmd.CombinedOutput()

		// Handle the specific exit code for service not found
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 113 {
				// Service not found or not installed
				return "[NOT INSTALLED]", nil
			}
		}

		// Handle other errors
		if err != nil {
			return "[FAILED]", fmt.Errorf("launchctl list failed: %v\nOutput: %s", err, string(output))
		}

		if strings.Contains(string(output), "com.facebook.osqueryd") {
			return "[RUNNING]", nil
		}
		return "[STOPPED]", nil

	case "windows":
		cmd := exec.Command("sc", "query", "osqueryd")
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

// OsqueryStart starts Osquery using platform-specific commands
func OsqueryStart() error {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("systemctl", "start", "osqueryd")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("systemctl start failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Osquery started successfully.")
		return nil

	case "darwin":
		cmd := exec.Command("sudo", "launchctl", "load", "/Library/LaunchDaemons/com.facebook.osqueryd.plist")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("launchctl load failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Osquery started successfully.")
		return nil

	case "windows":
		cmd := exec.Command("sc", "start", "osqueryd")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("sc start failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Osquery started successfully.")
		return nil

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// OsqueryStop stops Osquery using platform-specific commands
func OsqueryStop() error {
	switch runtime.GOOS {
	case "linux":
		cmd := exec.Command("systemctl", "stop", "osqueryd")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("systemctl stop failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Osquery stopped successfully.")
		return nil

	case "darwin":
		cmd := exec.Command("sudo", "launchctl", "unload", "/Library/LaunchDaemons/com.facebook.osqueryd.plist")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("launchctl unload failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Osquery stopped successfully.")
		return nil

	case "windows":
		cmd := exec.Command("sc", "stop", "osqueryd")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("sc stop failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Osquery stopped successfully.")
		return nil

	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func OsqueryRestart() error {
	// Stop and start Osquery
	err := OsqueryStop()
	if err != nil {
		return err
	}
	return OsqueryStart()
}
