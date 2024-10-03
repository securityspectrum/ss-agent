package zeek

import (
	"fmt"
	"os/exec"
	"ss-agent/utils/zeek"
	"strings"
)

// ZeekStatus checks the status of Zeek using `zeekctl status`
func ZeekStatus() (string, error) {
	fmt.Println("Checking Zeek status...")
	zeekctlPath, err := zeek.FindZeekctl()
	if err != nil {
		return "[NOT INSTALLED]", fmt.Errorf("failed to locate zeekctl: %v", err)
	}

	cmd := exec.Command(zeekctlPath, "status")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Handle the "stopped" case even if there is an error
		if strings.Contains(string(output), "stopped") {
			return "[STOPPED]", nil
		}
		return "[FAILED]", fmt.Errorf("zeekctl status failed: %v\nOutput: %s", err, string(output))
	}

	// Check for running or stopped
	if strings.Contains(string(output), "running") {
		return "[RUNNING]", nil
	} else if strings.Contains(string(output), "stopped") {
		return "[STOPPED]", nil
	}

	return "[UNKNOWN]", nil
}

// ZeekStart starts Zeek using `zeekctl deploy`
func ZeekStart() error {
	zeekctlPath, err := zeek.FindZeekctl()
	if err != nil {
		return fmt.Errorf("failed to locate zeekctl: %v", err)
	}

	cmd := exec.Command(zeekctlPath, "deploy")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zeekctl deploy failed: %v\nOutput: %s", err, string(output))
	}

	fmt.Println("Zeek started successfully.")
	return nil
}

// ZeekStop stops Zeek using `zeekctl stop`
func ZeekStop() error {
	zeekctlPath, err := zeek.FindZeekctl()
	if err != nil {
		return fmt.Errorf("failed to locate zeekctl: %v", err)
	}

	cmd := exec.Command(zeekctlPath, "stop")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zeekctl stop failed: %v\nOutput: %s", err, string(output))
	}

	fmt.Println("Zeek stopped successfully.")
	return nil
}

func ZeekRestart() error {
	zeekctlPath, err := zeek.FindZeekctl()
	if err != nil {
		return fmt.Errorf("failed to locate zeekctl: %v", err)
	}

	cmd := exec.Command(zeekctlPath, "restart")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("zeekctl restart failed: %v\nOutput: %s", err, string(output))
	}

	fmt.Println("Zeek restarted successfully.")
	return nil
}
