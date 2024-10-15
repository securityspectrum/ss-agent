package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func CreatePIDFile(pidFilePath, serviceName string) error {
	pidCmd := fmt.Sprintf("pgrep -f %s", serviceName)
	cmd := exec.Command("sh", "-c", pidCmd)
	pidOutput, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("could not find PID for service: %v", err)
	}
	pid := strings.TrimSpace(string(pidOutput))

	err = ioutil.WriteFile(pidFilePath, []byte(pid), 0644)
	if err != nil {
		return fmt.Errorf("could not write PID file: %v", err)
	}
	fmt.Printf("Created PID file for %s with PID %s.\n", serviceName, pid)
	return nil
}

func CheckProcessRunning(pid string) bool {
	pidInt, err := strconv.Atoi(strings.TrimSpace(pid))
	if err != nil {
		return false
	}
	processCmd := fmt.Sprintf("ps -p %d", pidInt)
	cmd := exec.Command("sh", "-c", processCmd)
	err = cmd.Run()
	return err == nil
}

func CreateDirectoryIfNotExists(dirPath string) error {
	// Check if the directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Directory does not exist, attempt to create it
		err := os.MkdirAll(dirPath, 0755) // 0755 permissions: owner can read/write/execute, others can read/execute
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dirPath, err)
		}
	}
	return nil
}
