// cmd.go
package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"ss-agent/api"
	"ss-agent/config"
	"ss-agent/service"
	"ss-agent/utils/osinfo"
)

var (
	configPath string // Holds the value of the --config flag
	debugMode  bool   // Holds the value of the --debug flag
)

const pidFile = "/tmp/ss-agent.pid" // Or use a directory within the user's home directory

func printConfig(cfg interface{}) {
	v := reflect.ValueOf(cfg)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()
		log.Printf("  %s: %v\n", field.Name, value)
	}
}

// Execute sets up and runs the Cobra command structure
func Execute(version string, ctx context.Context, cancel context.CancelFunc) {
	var client *http.Client

	rootCmd := &cobra.Command{
		Use:   "agent",
		Short: "SIEM Agent",
	}

	// Add global flags
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to configuration file")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug mode")

	// Define loadConfig as a PreRun hook
	loadConfig := func(cmd *cobra.Command, args []string) {
		// Load the config file if the --config flag is provided
		if configPath != "" {
			if err := config.LoadConfigFromFile(configPath); err != nil {
				log.Fatalf("Failed to load config from file: %v", err)
			}
		} else {
			// If no config path is provided, try default config paths
			if err := config.LoadConfig(); err != nil {
				log.Fatalf("Failed to load config from default paths: %v", err)
			}
		}

		// Setup TLS configuration with skip_ssl_verify option
		conf := config.GetConfig()
		tlsConfig := &tls.Config{
			InsecureSkipVerify: conf.SkipSSLVerify,
		}

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}

		// If debug mode is enabled, set more verbose logging and display config
		if debugMode {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Println("Configuration loaded:")
			printConfig(conf)
			log.Printf("OS Type: %s\n", osinfo.GetOSType())
			log.Printf("OS Distribution: %s\n", osinfo.GetOSDist())
			log.Printf("Program Version: %s\n", version)
		}
	}

	// runPingInIntervals function
	runPingInIntervals := func(ctx context.Context) {
		conf := config.GetConfig()
		pingInterval := time.Duration(conf.PingInterval) * time.Second
		if conf.PingInterval < 5 {
			pingInterval = 5 * time.Second
		}

		ticker := time.NewTicker(pingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping pinging due to service shutdown")
				return
			case <-ticker.C:
				if err := api.Ping(client); err != nil {
					log.Printf("ping failed: %v", err)
				}
			}
		}
	}

	// Start Command
	startCmd := &cobra.Command{
		Use:    "start",
		Short:  "Start the agent service",
		PreRun: loadConfig,
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Starting agent service...")
			writePidFile()
			go runPingInIntervals(ctx)
			<-ctx.Done() // Wait for context to be done
		},
	}

	// Daemon Command
	daemonCmd := &cobra.Command{
		Use:    "daemon",
		Short:  "Start the agent service in the background",
		PreRun: loadConfig,
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Starting agent service in the background...")
			cmdDaemonize()
		},
	}

	// Stop Command
	stopCmd := &cobra.Command{
		Use:    "stop",
		Short:  "Stop the agent service",
		PreRun: loadConfig,
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Stopping agent service...")
			stopService()
		},
	}

	// Status Command (no PreRun)
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Get the agent status",
		Run: func(cmd *cobra.Command, args []string) {
			if debugMode {
				log.Println("Getting agent status in debug mode...")
				log.SetFlags(log.LstdFlags | log.Lshortfile) // More verbose logging
			} else {
				log.Println("Getting agent status...")
				log.SetFlags(log.LstdFlags)
			}
			statusService()
		},
	}

	// Register Command
	registerCmd := &cobra.Command{
		Use:    "register",
		Short:  "Register the agent",
		PreRun: loadConfig,
		Run: func(cmd *cobra.Command, args []string) {
			if debugMode {
				log.Println("Registering agent in debug mode...")
				log.SetFlags(log.LstdFlags | log.Lshortfile) // More verbose logging
			} else {
				log.Println("Registering agent...")
				log.SetFlags(log.LstdFlags)
			}
			if err := api.RegisterAgent(); err != nil {
				log.Fatalf("Failed to register agent: %v", err)
			}
		},
	}

	// Unregister Command
	unregisterCmd := &cobra.Command{
		Use:    "unregister",
		Short:  "Unregister the agent",
		PreRun: loadConfig,
		Run: func(cmd *cobra.Command, args []string) {
			if debugMode {
				log.Println("Unregistering agent in debug mode...")
				log.SetFlags(log.LstdFlags | log.Lshortfile) // More verbose logging
			} else {
				log.Println("Unregistering agent...")
				log.SetFlags(log.LstdFlags)
			}
			if err := api.UnregisterAgent(); err != nil {
				log.Fatalf("Failed to unregister agent: %v", err)
			}
		},
	}

	// Ping Command
	pingCmd := &cobra.Command{
		Use:    "ping",
		Short:  "Ping the SIEM server once",
		PreRun: loadConfig,
		Run: func(cmd *cobra.Command, args []string) {
			if debugMode {
				log.Println("Pinging SIEM server in debug mode...")
				log.SetFlags(log.LstdFlags | log.Lshortfile) // More verbose logging
			} else {
				log.Println("Pinging SIEM server...")
				log.SetFlags(log.LstdFlags)
			}
			if err := api.Ping(client); err != nil {
				log.Fatalf("Failed to ping SIEM server: %v", err)
			}
		},
	}

	// Version Command (No PreRun)
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of the agent",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("SIEM Agent version: %s\n", version)
		},
	}

	// Service Command
	serviceCmd := &cobra.Command{
		Use:   "service",
		Short: "Manage services (install, uninstall, start, stop, restart, status)",
	}

	// Install Service Command
	installCmd := &cobra.Command{
		Use:    "install [service]",
		Short:  "Install a service (fluent-bit, zeek, osquery)",
		PreRun: loadConfig,
		Args:   cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if debugMode {
				log.Printf("Installing service %s in debug mode...", args[0])
				log.SetFlags(log.LstdFlags | log.Lshortfile) // More verbose logging
			} else {
				log.Printf("Installing service %s...", args[0])
				log.SetFlags(log.LstdFlags)
			}
			if err := service.InstallService(args[0]); err != nil {
				log.Fatalf("Failed to install service: %v", err)
			}
		},
	}

	// Uninstall Service Command
	uninstallCmd := &cobra.Command{
		Use:    "uninstall [service]",
		Short:  "Uninstall a service (fluent-bit, zeek, osquery)",
		PreRun: loadConfig,
		Args:   cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if debugMode {
				log.Printf("Uninstalling service %s in debug mode...", args[0])
				log.SetFlags(log.LstdFlags | log.Lshortfile) // More verbose logging
			} else {
				log.Printf("Uninstalling service %s...", args[0])
				log.SetFlags(log.LstdFlags)
			}
			if err := service.UninstallService(args[0]); err != nil {
				log.Fatalf("Failed to uninstall service: %v", err)
			}
		},
	}

	// Add install and uninstall to service command
	serviceCmd.AddCommand(installCmd, uninstallCmd)

	// Add all commands to rootCmd
	rootCmd.AddCommand(startCmd, daemonCmd, stopCmd, statusCmd, registerCmd, unregisterCmd, pingCmd, versionCmd, serviceCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

// cmdDaemonize starts the agent as a background process
func cmdDaemonize() {
	cmd := exec.Command(os.Args[0], "start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	fmt.Printf("Agent service started in the background with PID %d\n", cmd.Process.Pid)
	os.Exit(0)
}

// writePidFile writes the current process PID to the pidFile
func writePidFile() {
	const pidFile = "/tmp/ss-agent.pid"
	pid := os.Getpid()
	err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		log.Fatalf("Failed to write PID file: %v", err)
	}
}

// stopService sends a SIGTERM to the process whose PID is stored in pidFile
func stopService() {
	const pidFile = "/tmp/ss-agent.pid"
	data, err := os.ReadFile(pidFile)
	if err != nil {
		log.Fatalf("Failed to read PID file: %v", err)
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		log.Fatalf("Invalid PID in PID file: %v", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		log.Fatalf("Failed to find process with PID %d: %v", pid, err)
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		log.Fatalf("Failed to send SIGTERM to process with PID %d: %v", pid, err)
	}

	fmt.Printf("Sent SIGTERM to process with PID %d\n", pid)
}

// statusService checks if the process with PID from pidFile is running
func statusService() {
	const pidFile = "/tmp/ss-agent.pid"
	data, err := os.ReadFile(pidFile)
	if err != nil {
		fmt.Println("stopped")
		return
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		fmt.Println("stopped")
		return
	}

	process, err := os.FindProcess(pid)
	if err != nil || process.Signal(syscall.Signal(0)) != nil {
		fmt.Println("stopped")
	} else {
		fmt.Println("running")
	}
}
