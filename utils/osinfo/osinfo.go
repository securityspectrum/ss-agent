package osinfo

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"ss-agent/config"
)

func getOSReleaseInfo() (map[string]string, error) {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, fmt.Errorf("error opening /etc/os-release: %v", err)
	}
	defer file.Close()

	info := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := strings.Trim(parts[1], `"`)
		info[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading /etc/os-release: %v", err)
	}

	return info, nil
}

func DetectOS() {
	switch runtime.GOOS {
	case "windows":
		config.SetOS("windows", "windows")
	case "darwin":
		config.SetOS("darwin", "macos")
	case "linux":
		info, err := getOSReleaseInfo()
		if err != nil {
			log.Fatalf("Failed to detect Linux distribution: %v", err)
		}
		dist := info["ID"]
		version := info["VERSION"]
		config.SetOS("linux", dist+" "+version)
	default:
		log.Fatalf("Unsupported operating system: %s", runtime.GOOS)
	}
}
