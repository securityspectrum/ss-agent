package osinfo

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

var osType string
var osDist string

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
		osType = "windows"
		osDist = "windows"
	case "darwin":
		osType = "darwin"
		osDist = "macos"
	case "linux":
		info, err := getOSReleaseInfo()
		if err != nil {
			log.Fatalf("Failed to detect Linux distribution: %v", err)
		}
		osType = "linux"
		osDist = normalizeLinuxDist(info["ID"])
	default:
		log.Fatalf("Unsupported operating system: %s", runtime.GOOS)
	}
}

func normalizeLinuxDist(dist string) string {
	switch dist {
	case "ubuntu", "debian", "mint":
		return dist
	case "fedora", "rhel", "centos":
		return dist
	default:
		return "unsupported"
	}
}

func GetOSType() string {
	return osType
}

func GetOSDist() string {
	return osDist
}
