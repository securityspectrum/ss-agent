// cmd/cmd.go

package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"ss-agent/api"
	"ss-agent/config"
	"ss-agent/service"
	"ss-agent/utils"
	"ss-agent/utils/osinfo"
)

var (
	client     *http.Client
	configPath string // Holds the value of the --config flag
	debugMode  bool   // Holds the value of the --debug flag
	daemonMode bool   // Holds the value of the --daemon flag
)

// const pidFile = "/tmp/ss-agent.pid" // Or use a directory within the user's home directory

var (
	pidFile string // Declared as a variable instead of a constant
)

// Define valid services including 'all'
var validServices = []string{"fluent-bit", "zeek", "osqueryd", "all"}

// Helper function to check if a slice contains a string (case-insensitive)
func contains(slice []string, item string) bool {
	item = strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == item {
			return true
		}
	}
	return false
}

// printConfig logs the current configuration in a readable format
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
	var rootCmd = &cobra.Command{
		Use:   "agent",
		Short: "SIEM Agent",
	}

	pidFile = getPidFilePath()
	log.Printf("PID file will be created at: %s", pidFile)
	// Ensure the PID directory exists before trying to write to it
	err := utils.CreateDirectoryIfNotExists(filepath.Dir(pidFile))
	if err != nil {
		log.Fatalf("Failed to create directory for PID file: %v", err)
	}

	// Add global flags
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "Path to configuration file")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "Enable debug mode")

	osinfo.DetectOS()

	// loadConfig is a PreRun hook that loads the configuration based on the --config flag
	loadConfig := func(cmd *cobra.Command, args []string) {
		// Load the config file if the --config flag is provided
		if configPath != "" {
			log.Printf("Loading configuration from: %s\n", configPath)
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
			printConfig(config.GetConfig())
			log.Printf("OS Type: %s\n", osinfo.GetOSType())
			log.Printf("OS Distribution: %s\n", osinfo.GetOSDist())
			log.Printf("Program Version: %s\n", version)
		}
	}

	// Check if an instance is already running
	checkRunningInstance := func() bool {
		data, err := os.ReadFile(pidFile)
		if err == nil { // If PID file exists
			pid, err := strconv.Atoi(string(data))
			if err == nil {
				process, err := os.FindProcess(pid)
				if err == nil && process.Signal(syscall.Signal(0)) == nil {
					// Process is running
					log.Printf("An instance of the agent is already running with PID %d\n", pid)
					return true
				}
			}
		}
		// No running instance found or PID file does not exist
		return false
	}

	// Start Command
	var startCmd = &cobra.Command{
		Use:    "start",
		Short:  "Start the agent service",
		PreRun: loadConfig,
		Run: func(cmd *cobra.Command, args []string) {
			// Prevent starting a new instance if one is already running
			if checkRunningInstance() {
				log.Fatal("Failed to start: another instance of the agent is already running.")
			}

			if daemonMode {
				// If daemon mode is enabled, start in the background
				log.Println("Starting agent service in the background...")
				cmdDaemonize()
			} else {
				// Otherwise, start normally
				log.Println("Starting agent service...")
				writePidFile()
				go runPingInIntervals(ctx)
				<-ctx.Done() // Wait for context to be done
			}
		},
	}

	// Stop Command
	var stopCmd = &cobra.Command{
		Use:    "stop",
		Short:  "Stop the agent service",
		PreRun: loadConfig,
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Stopping agent service...")
			stopService()
		},
	}

	// Status Command
	var statusCmd = &cobra.Command{
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
	var registerCmd = &cobra.Command{
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
	var unregisterCmd = &cobra.Command{
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
	var pingCmd = &cobra.Command{
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

	// Version Command (No PreRun, hence no config loading)
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of the agent",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("SIEM Agent version: %s\n", version)
		},
	}

	// Service Command
	var serviceCmd = &cobra.Command{
		Use:   "service",
		Short: "Manage services (start, stop, restart, status)",
	}

	// Start Service Command
	var serviceStartCmd = &cobra.Command{
		Use:    "start [service|all]",
		Short:  "Start a service (fluent-bit, zeek, osqueryd) or all services",
		PreRun: loadConfig,
		Args:   cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			serviceName := strings.ToLower(args[0])
			if serviceName == "all" {
				log.Println("Starting all services...")
				for _, svc := range service.AllServices {
					log.Printf("Attempting to start %s...", svc)
					if err := service.ManageService(svc, "start"); err != nil {
						log.Printf("Failed to start %s: %v", svc, err)
						fmt.Printf("%-15s: [FAILED] %v\n", svc, err)
					} else {
						log.Printf("Successfully started %s", svc)
						fmt.Printf("%-15s: [STARTED]\n", svc)
					}
				}
			} else if contains(service.AllServices, serviceName) {
				log.Printf("Starting service %s...", serviceName)
				if err := service.ManageService(serviceName, "start"); err != nil {
					log.Fatalf("Failed to start service %s: %v", serviceName, err)
				} else {
					log.Printf("Service %s started successfully", serviceName)
					fmt.Println("Service started successfully")
				}
			} else {
				fmt.Printf("Invalid service name: %s\n", args[0])
				fmt.Printf("Valid services are: %s\n", strings.Join(service.AllServices, ", "))
				os.Exit(1)
			}
		},
	}

	// Stop Service Command
	var serviceStopCmd = &cobra.Command{
		Use:    "stop [service|all]",
		Short:  "Stop a service (fluent-bit, zeek, osqueryd) or all services",
		PreRun: loadConfig,
		Args:   cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			serviceName := strings.ToLower(args[0])
			if serviceName == "all" {
				log.Println("Stopping all services...")
				for _, svc := range service.AllServices {
					log.Printf("Attempting to stop %s...", svc)
					if err := service.ManageService(svc, "stop"); err != nil {
						log.Printf("Failed to stop %s: %v", svc, err)
						fmt.Printf("%-15s: [FAILED] %v\n", svc, err)
					} else {
						log.Printf("Successfully stopped %s", svc)
						fmt.Printf("%-15s: [STOPPED]\n", svc)
					}
				}
			} else if contains(service.AllServices, serviceName) {
				log.Printf("Stopping service %s...", serviceName)
				if err := service.ManageService(serviceName, "stop"); err != nil {
					log.Fatalf("Failed to stop service %s: %v", serviceName, err)
				} else {
					log.Printf("Service %s stopped successfully", serviceName)
					fmt.Println("Service stopped successfully")
				}
			} else {
				fmt.Printf("Invalid service name: %s\n", args[0])
				fmt.Printf("Valid services are: %s\n", strings.Join(service.AllServices, ", "))
				os.Exit(1)
			}
		},
	}

	// Restart Service Command
	var serviceRestartCmd = &cobra.Command{
		Use:    "restart [service|all]",
		Short:  "Restart a service (fluent-bit, zeek, osqueryd) or all services",
		PreRun: loadConfig,
		Args:   cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			serviceName := strings.ToLower(args[0])
			if serviceName == "all" {
				log.Println("Restarting all services...")
				for _, svc := range service.AllServices {
					log.Printf("Attempting to restart %s...", svc)
					if err := service.ManageService(svc, "restart"); err != nil {
						log.Printf("Failed to restart %s: %v", svc, err)
						fmt.Printf("%-15s: [FAILED] %v\n", svc, err)
					} else {
						log.Printf("Successfully restarted %s", svc)
						fmt.Printf("%-15s: [RESTARTED]\n", svc)
					}
				}
			} else if contains(service.AllServices, serviceName) {
				log.Printf("Restarting service %s...", serviceName)
				if err := service.ManageService(serviceName, "restart"); err != nil {
					log.Fatalf("Failed to restart service %s: %v", serviceName, err)
				} else {
					log.Printf("Service %s restarted successfully", serviceName)
					fmt.Println("Service restarted successfully")
				}
			} else {
				fmt.Printf("Invalid service name: %s\n", args[0])
				fmt.Printf("Valid services are: %s\n", strings.Join(service.AllServices, ", "))
				os.Exit(1)
			}
		},
	}

	// Status Service Command
	var serviceStatusCmd = &cobra.Command{
		Use:   "status [service|all]",
		Short: "Get the status of a service (fluent-bit, zeek, osqueryd) or all services",
		Long: `Check the status of a specific service or all managed services.

Examples:
  ss-agent service status zeek
  ss-agent service status all`,
		PreRun: loadConfig,
		Args:   cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			serviceName := strings.ToLower(args[0])
			if !contains(validServices, serviceName) {
				fmt.Printf("Invalid service name: %s\n", args[0])
				fmt.Printf("Valid services are: %s\n", strings.Join(service.AllServices, ", "))
				fmt.Printf("Or use 'all' to manage all services.\n")
				os.Exit(1)
			}
			if debugMode {
				log.Printf("Checking status of service %s in debug mode...", serviceName)
				log.SetFlags(log.LstdFlags | log.Lshortfile)
			} else {
				log.Printf("Checking status of service %s...", serviceName)
				log.SetFlags(log.LstdFlags)
			}
			service.HealthCheck(serviceName)
		},
	}

	// Add subcommands to 'service' command
	serviceCmd.AddCommand(serviceStartCmd, serviceStopCmd, serviceRestartCmd, serviceStatusCmd)

	// Add 'service' and other commands to root
	rootCmd.AddCommand(startCmd, stopCmd, statusCmd, registerCmd, unregisterCmd, pingCmd, versionCmd, serviceCmd)

	// Add daemon flag to start command
	startCmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "Run the agent service in the background")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

// ensureLogDirectory ensures that the log directory exists and creates it if necessary
func ensureLogDirectory(logFilePath string) error {
	log.Println("Ensuring log directory exists: ", logFilePath)
	logDir := filepath.Dir(logFilePath)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		err := os.MkdirAll(logDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create log directory %s: %v", logDir, err)
		}
	}
	return nil
}

func getLogFilePath() string {
	ostype := osinfo.GetOSType()

	log.Printf("Detected OS: %s\n", ostype)
	switch ostype {
	case "linux":
		return "/var/log/ss-agent/ss-agent.log"
	case "windows":
		return `C:\ProgramData\ss-agent\ss-agent.log`
	case "darwin": // macOS
		return "/Library/Logs/ss-agent/ss-agent.log"
	default:
		// Default fallback if OS detection fails
		return "./ss-agent.log"
	}
}

// cmdDaemonize starts the agent as a background process
func cmdDaemonize() {
	// Prepare the command to rerun the current binary in "start" mode
	cmd := exec.Command(os.Args[0], "start", "--config", configPath)

	// Get the log file path
	logFilePath := getLogFilePath()

	// Ensure log directory exists
	if err := ensureLogDirectory(logFilePath); err != nil {
		log.Fatalf("Failed to ensure log directory: %v", err)
	}

	// Open or create the log file to write logs in daemon mode
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Redirect stdout and stderr to the log file in daemon mode
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Start the process in the background
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start daemon process: %v", err)
	}

	// Output the PID to the log file instead of the console
	log.Printf("Agent service started in the background with PID %d\n", cmd.Process.Pid)

	// Exit the parent process to complete the daemonization
	os.Exit(0)
}

func getPidFilePath() string {
	ostype := osinfo.GetOSType()
	switch ostype {
	case "windows":
		return `C:/ProgramData/ss-agent/ss-agent.pid`
	case "darwin": // macOS
		return "/Library/Logs/ss-agent/ss-agent.pid"
	case "linux":
		return "/var/run/ss-agent.pid"
	default:
		tempDir := os.TempDir()
		return filepath.Join(tempDir, "ss-agent.pid")
	}
}

// writePidFile writes the current process PID to the pidFile
func writePidFile() {
	pid := os.Getpid()
	err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		log.Fatalf("Failed to write PID file: %v", err)
	}
}

// stopService sends a SIGTERM to the process whose PID is stored in pidFile
func stopService() {
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

// runPingInIntervals continuously pings the SIEM server based on the PingInterval
func runPingInIntervals(ctx context.Context) {
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
