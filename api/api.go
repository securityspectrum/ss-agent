package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"ss-agent/config"
)

func RegisterAgent() error {
	log.Println("Registering agent...")
	// Implement registration logic
	return nil
}

func UnregisterAgent() error {
	log.Println("Unregistering agent...")
	// Implement unregistration logic
	return nil
}

func Ping(client *http.Client) error {
	conf := config.GetConfig()
	if conf.APIUrl == "" {
		return fmt.Errorf("APIUrl is not set in the configuration")
	}

	url := fmt.Sprintf("%s/agents/ping", conf.APIUrl)
	log.Printf("ping %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	setHeaders(req, conf)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ping failed: %v", err)
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ping failed: %v", err)
		return fmt.Errorf("failed to read response body: %v", err)
	}

	log.Printf("pong: %s", body)
	return nil
}

func Status() {
	log.Println("Agent status: Registered")
	// Implement status logic
}

func setHeaders(req *http.Request, conf config.Config) {
	req.Header.Set("X-ORGANIZATION-KEY", conf.OrganizationKey)
	req.Header.Set("X-API-ACCESS-KEY", conf.APIAccessKey)
	req.Header.Set("X-API-SECRET-KEY", conf.APISecretKey)
}
