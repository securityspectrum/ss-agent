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
	OS              string `json:"os"`
	Dist            string `json:"dist"`
	PingInterval    int    `json:"ping_interval"`
}

var config Config

func LoadConfig() {
	file, err := os.Open("config/config.json")
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

func SetOS(os, dist string) {
	config.OS = os
	config.Dist = dist
}
