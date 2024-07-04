package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	APIUrl          string `json:"api_url"`
	OrganizationKey string `json:"organization_key"`
	APIAccessKey    string `json:"api_access_key"`
	APISecretKey    string `json:"api_secret_key"`
	CertFile        string `json:"cert_file"`
	KeyFile         string `json:"key_file"`
	CAFile          string `json:"ca_file"`
	PingInterval    int    `json:"ping_interval"`
}

var config Config

var defaultConfigPaths = []string{
	"./config/config.json",
	"/etc/ss-agent/config/config.json",
	"/usr/local/etc/ss-agent/config/config.json",
	"/usr/local/ss-agent/config/config.json",
	"C:\\ProgramData\\ss-agent\\config\\config.json",
}

func findConfigFile() (string, error) {
	for _, path := range defaultConfigPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", os.ErrNotExist
}

func LoadConfig() {
	configPath, err := findConfigFile()
	if err != nil {
		log.Fatalf("Error finding config file: %v", err)
	}

	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Error decoding config file: %v", err)
	}

	// Set default ping interval if not set or invalid
	if config.PingInterval < 5 {
		config.PingInterval = 5
	}
}

func GetConfig() Config {
	return config
}
