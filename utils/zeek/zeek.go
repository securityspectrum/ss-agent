package zeek

import (
	"fmt"
	"os"
	"os/exec"
)

func FindZeekctl() (string, error) {
	// Log information
	fmt.Println("Searching for zeekctl...")

	// List of possible paths where zeekctl might be located
	possiblePaths := []string{
		"/usr/bin/zeekctl",
		"/opt/zeek/bin/zeekctl",
		"/usr/local/bin/zeekctl",
		"/usr/local/opt/zeek/bin/zeekctl",
	}

	// Try each of the possible paths
	for _, path := range possiblePaths {
		fmt.Printf("Checking path: %s\n", path)
		if _, err := os.Stat(path); err == nil {
			// File exists, zeekctl found
			fmt.Printf("Found zeekctl at: %s\n", path)
			return path, nil
		}
	}

	// If not found, try to find it in the system PATH using LookPath
	zeekctlPath, err := exec.LookPath("zeekctl")
	if err == nil {
		fmt.Printf("Found zeekctl in PATH: %s\n", zeekctlPath)
		return zeekctlPath, nil
	}

	// If zeekctl was not found, return an error
	return "", fmt.Errorf("zeekctl executable not found in known paths or system PATH")
}
