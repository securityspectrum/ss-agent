package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"ss-agent/cmd"
	"ss-agent/config"
	"ss-agent/utils/osinfo"
	"ss-agent/utils/tlsconfig"
)

var version = "1.0.0"

func printConfig(cfg interface{}) {
	v := reflect.ValueOf(cfg)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()
		log.Printf("  %s: %v\n", field.Name, value)
	}
}

func main() {
	// Check if the application is running in debug mode
	debugMode := os.Getenv("DEBUG_MODE") == "true"

	// Configure logging to include timestamps
	log.SetFlags(log.LstdFlags)

	// Load configuration
	config.LoadConfig()
	osinfo.DetectOS()

	// Display loaded configurations and versions if in debug mode
	if debugMode {
		log.Println("Loaded Configurations:")
		printConfig(config.GetConfig())
		log.Printf("OS Version: %s\n", config.GetConfig().Dist)
		log.Printf("Program Version: %s\n", version)
	}

	tlsConfig, err := tlsconfig.SetupTLSConfig()
	if err != nil {
		log.Fatalf("Failed to setup TLS config: %v", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	// Channel to listen for termination signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-stop
		log.Println("Shutting down gracefully...")

		ctxShutdown, cancelShutdown := context.WithTimeout(ctx, 5*time.Second)
		defer cancelShutdown()

		if err := cmd.Shutdown(ctxShutdown, client); err != nil {
			log.Fatalf("Error during shutdown: %v", err)
		}

		log.Println("Service stopped.")
		cancel()
	}()

	cmd.Execute(client, version, ctx, cancel)
}
