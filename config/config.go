// config.go
package config

import (
	"encoding/json"
	"errors"
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
	SkipSSLVerify   bool   `json:"skip_ssl_verify"`
}

var config Config

var defaultConfigPaths = []string{
	"./config/config.json",
	"/etc/ss-agent/config/config.json",
	"/usr/local/etc/ss-agent/config/config.json",
	"/usr/local/ss-agent/config/config.json",
	"C:\\ProgramData\\ss-agent\\config\\config.json",
}

// findConfigFile searches for the configuration file in default paths
func findConfigFile() (string, error) {
	for _, path := range defaultConfigPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", os.ErrNotExist
}

// LoadConfigFromFile loads configuration from the specified file path
func LoadConfigFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return err
	}

	// Set default ping interval if not set or invalid
	if config.PingInterval < 5 {
		config.PingInterval = 5
	}

	// No need to set SkipSSLVerify to false explicitly since it's already false by default

	log.Println("Configuration loaded from:", filePath)
	return nil
}

// LoadConfig attempts to load configuration from default paths
func LoadConfig() error {
	configPath, err := findConfigFile()
	if err != nil {
		return errors.New("no config file found in default paths")
	}
	return LoadConfigFromFile(configPath)
}

// GetConfig returns the current configuration
func GetConfig() Config {
	return config
}
