package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"ss-agent/api"
	"ss-agent/config"
	"ss-agent/service"
)

var client *http.Client
var cancelFunc context.CancelFunc

const pidFile = "/tmp/ss-agent.pid" // Or use a directory within the user's home directory

func setClient(httpClient *http.Client) {
	client = httpClient
}

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

func Shutdown(ctx context.Context, client *http.Client) error {
	// Add any cleanup logic here

	// Example: Closing any remaining HTTP connections
	if client != nil && client.Transport != nil {
		if transport, ok := client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}

	// Wait for context timeout or completion
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return nil
	}
}

func Execute(httpClient *http.Client, version string, ctx context.Context, cancel context.CancelFunc) {
	setClient(httpClient)
	cancelFunc = cancel

	var rootCmd = &cobra.Command{
		Use:   "agent",
		Short: "SIEM Agent",
	}

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the agent service",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Starting agent service...")
			writePidFile()
			go runPingInIntervals(ctx)
			<-ctx.Done() // Wait for context to be done
		},
	}

	var daemonCmd = &cobra.Command{
		Use:   "start -d",
		Short: "Start the agent service in the background",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Starting agent service in the background...")
			cmdDaemonize()
		},
	}

	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop the agent service",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Stopping agent service...")
			stopService()
		},
	}

	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Get the agent status",
		Run: func(cmd *cobra.Command, args []string) {
			statusService()
		},
	}

	var registerCmd = &cobra.Command{
		Use:   "register",
		Short: "Register the agent",
		Run: func(cmd *cobra.Command, args []string) {
			if err := api.RegisterAgent(); err != nil {
				log.Fatalf("Failed to register agent: %v", err)
			}
		},
	}

	var unregisterCmd = &cobra.Command{
		Use:   "unregister",
		Short: "Unregister the agent",
		Run: func(cmd *cobra.Command, args []string) {
			if err := api.UnregisterAgent(); err != nil {
				log.Fatalf("Failed to unregister agent: %v", err)
			}
		},
	}

	var pingCmd = &cobra.Command{
		Use:   "ping",
		Short: "Ping the SIEM server once",
		Run: func(cmd *cobra.Command, args []string) {
			if err := api.Ping(client); err != nil {
				log.Fatalf("Failed to ping SIEM server: %v", err)
			}
		},
	}

	var serviceCmd = &cobra.Command{
		Use:   "service",
		Short: "Manage services (install, uninstall, start, stop, restart, status)",
	}

	var installCmd = &cobra.Command{
		Use:   "install [service]",
		Short: "Install a service (fluent-bit, zeek, osquery)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := service.InstallService(args[0]); err != nil {
				log.Fatalf("Failed to install service: %v", err)
			}
		},
	}

	var uninstallCmd = &cobra.Command{
		Use:   "uninstall [service]",
		Short: "Uninstall a service (fluent-bit, zeek, osquery)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := service.UninstallService(args[0]); err != nil {
				log.Fatalf("Failed to uninstall service: %v", err)
			}
		},
	}

	serviceCmd.AddCommand(installCmd, uninstallCmd)
	rootCmd.AddCommand(startCmd, daemonCmd, stopCmd, statusCmd, registerCmd, unregisterCmd, pingCmd, serviceCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

func cmdDaemonize() {
	cmd := exec.Command(os.Args[0], "start")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	fmt.Printf("Agent service started in the background with PID %d\n", cmd.Process.Pid)
	os.Exit(0)
}

func writePidFile() {
	pid := os.Getpid()
	err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		log.Fatalf("Failed to write PID file: %v", err)
	}
}

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
