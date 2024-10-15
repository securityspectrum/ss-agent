package zeek

import (
	"fmt"
	"os/exec"
	"runtime"
	"ss-agent/utils/zeek"
	"strings"
)

// ZeekStatus checks the status of Zeek using `zeekctl status`
func ZeekStatus() (string, error) {
	fmt.Println("Checking Zeek status...")

	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("sc", "query", "ss-network-analyzer")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "[FAILED]", fmt.Errorf("sc query failed: %v\nOutput: %s", err, string(output))
		}

		if strings.Contains(string(output), "RUNNING") {
			return "[RUNNING]", nil
		}
		return "[STOPPED]", nil

	case "darwin", "linux":
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

		if strings.Contains(string(output), "running") {
			return "[RUNNING]", nil
		} else if strings.Contains(string(output), "stopped") {
			return "[STOPPED]", nil
		}
	default:
		return "[UNKNOWN]", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return "[UNKNOWN]", nil
}

// ZeekStart starts Zeek using `zeekctl deploy`
func ZeekStart() error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("sc", "start", "ss-network-analyzer")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("sc start failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Zeek started successfully.")
		return nil

	case "darwin", "linux":
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
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// ZeekStop stops Zeek using `zeekctl stop`
func ZeekStop() error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("sc", "stop", "ss-network-analyzer")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("sc stop failed: %v\nOutput: %s", err, string(output))
		}
		fmt.Println("Zeek stopped successfully.")
		return nil

	case "darwin", "linux":
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
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func ZeekRestart() error {
	err := ZeekStop()
	if err != nil {
		return err
	}
	return ZeekStart()
}
