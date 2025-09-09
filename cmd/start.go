package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/shiquda/lai/internal/collector"
	"github.com/shiquda/lai/internal/daemon"
	"github.com/shiquda/lai/internal/platform"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [log-file]",
	Short: "Start log monitoring",
	Long:  "Start log monitoring service for the specified log file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logFile := args[0]

		// Get command line parameters
		lineThreshold, _ := cmd.Flags().GetInt("line-threshold")
		intervalStr, _ := cmd.Flags().GetString("interval")
		daemonMode, _ := cmd.Flags().GetBool("daemon")
		processName, _ := cmd.Flags().GetString("name")
		errorOnlyMode, _ := cmd.Flags().GetBool("error-only")
		enabledNotifiers, _ := cmd.Flags().GetStringSlice("notifiers")

		// Parse time interval
		var checkInterval *time.Duration
		if intervalStr != "" {
			if duration, err := time.ParseDuration(intervalStr); err != nil {
				log.Fatalf("Invalid interval format: %v", err)
			} else {
				checkInterval = &duration
			}
		}

		// Handle line-threshold parameter
		var lineThresholdPtr *int
		if cmd.Flags().Changed("line-threshold") {
			lineThresholdPtr = &lineThreshold
		}

		// Handle error-only parameter
		var errorOnlyModePtr *bool
		if cmd.Flags().Changed("error-only") {
			errorOnlyModePtr = &errorOnlyMode
		}

		if daemonMode {
			if err := runDaemon(logFile, lineThresholdPtr, checkInterval, processName, errorOnlyModePtr, enabledNotifiers); err != nil {
				log.Fatalf("Daemon startup failed: %v", err)
			}
		} else {
			if err := runMonitor(logFile, lineThresholdPtr, checkInterval, errorOnlyModePtr, enabledNotifiers); err != nil {
				log.Fatalf("Monitor failed: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Add command line parameters
	startCmd.Flags().IntP("line-threshold", "l", 0, "Number of lines to trigger summary (overrides global config)")
	startCmd.Flags().StringP("interval", "i", "", "Check interval (e.g., 30s, 1m) (overrides global config)")
	startCmd.Flags().BoolP("daemon", "d", false, "Run in daemon mode (background)")
	startCmd.Flags().StringP("name", "n", "", "Custom process name (used in daemon mode)")
	startCmd.Flags().BoolP("error-only", "E", false, "Only send notifications for errors and exceptions")
	startCmd.Flags().StringSlice("notifiers", []string{}, "Enable specific notifiers (comma-separated: telegram,email)")
}

func runMonitor(logFile string, lineThreshold *int, checkInterval *time.Duration, errorOnlyMode *bool, enabledNotifiers []string) error {
	// Create file monitor source
	monitorSource := collector.NewFileSource(logFile)

	// Build unified configuration
	cfg, err := collector.BuildMonitorConfig(monitorSource, lineThreshold, checkInterval, nil, "", nil, errorOnlyMode, nil)
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Create unified monitor
	monitor, err := collector.NewUnifiedMonitor(cfg, enabledNotifiers)
	if err != nil {
		return fmt.Errorf("failed to create unified monitor: %w", err)
	}

	// Start monitoring
	return monitor.Start()
}

func runDaemon(logFile string, lineThreshold *int, checkInterval *time.Duration, processName string, errorOnlyMode *bool, enabledNotifiers []string) error {
	// Create daemon manager
	manager, err := daemon.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create daemon manager: %w", err)
	}

	// Generate process ID
	var processID string
	isResumeMode := os.Getenv("LAI_RESUME_MODE") == "1"

	if processName != "" {
		processID = manager.GenerateProcessIDWithName(processName)
		// Check if process name already exists (only in parent process and not in resume mode)
		if os.Getenv("LAI_DAEMON_MODE") != "1" && !isResumeMode && manager.ProcessExists(processID) {
			return fmt.Errorf("process with name '%s' already exists", processName)
		}
	} else {
		processID = manager.GenerateProcessID(logFile)
	}

	// Get log file path for daemon
	daemonLogPath := manager.GetProcessLogPath(processID)

	// Fork and create daemon process
	if os.Getenv("LAI_DAEMON_MODE") != "1" {
		// Parent process - start daemon
		os.Setenv("LAI_DAEMON_MODE", "1")

		// Redirect output to log file
		logFileHandle, err := os.OpenFile(daemonLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to create log file: %w", err)
		}
		defer logFileHandle.Close()

		// Start daemon process - get full executable path
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}
		cmd := execPath
		args := append([]string{cmd}, os.Args[1:]...)

		// Use platform-specific daemon process creation
		p := platform.New()
		process, err := p.Process.StartDaemonProcess(cmd, args, logFileHandle, append(os.Environ(), "LAI_DAEMON_MODE=1"))
		if err != nil {
			return fmt.Errorf("failed to start daemon process: %w", err)
		}

		// Save process information (only if not in resume mode)
		if !isResumeMode {
			processInfo := &daemon.ProcessInfo{
				ID:        processID,
				PID:       process.Pid,
				LogFile:   logFile,
				StartTime: time.Now(),
				Status:    "running",
			}

			if err := manager.SaveProcessInfo(processInfo); err != nil {
				return fmt.Errorf("failed to save process info: %w", err)
			}
		}

		fmt.Printf("Started daemon with process ID: %s (PID: %d)\n", processID, process.Pid)
		fmt.Printf("Log file: %s\n", daemonLogPath)
		fmt.Printf("Use 'lai list' to see running processes\n")
		fmt.Printf("Use 'lai logs %s' to view logs\n", processID)
		fmt.Printf("Use 'lai stop %s' to stop the process\n", processID)

		return nil
	}

	// Child process - run as daemon
	// First update our status to ensure we're properly registered
	if isResumeMode {
		// In resume mode, the process info should already exist with "resuming" status
		// We don't need to update it here as resume command will update it
	} else {
		// In normal start mode, ensure our process info is saved
		processInfo := &daemon.ProcessInfo{
			ID:        processID,
			PID:       os.Getpid(),
			LogFile:   logFile,
			StartTime: time.Now(),
			Status:    "running",
		}
		if err := manager.SaveProcessInfo(processInfo); err != nil {
			log.Printf("Failed to save process info in child: %v", err)
		}
	}

	// Setup cleanup for daemon process
	defer func() {
		// Update process status to stopped instead of removing
		info, err := manager.LoadProcessInfo(processID)
		if err == nil {
			info.Status = "stopped"
			if saveErr := manager.SaveProcessInfo(info); saveErr != nil {
				log.Printf("Failed to update process status: %v", saveErr)
			}
		}
	}()

	// Run the actual monitoring
	return runMonitor(logFile, lineThreshold, checkInterval, errorOnlyMode, enabledNotifiers)
}
