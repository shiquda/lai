package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shiquda/lai/internal/collector"
	"github.com/shiquda/lai/internal/daemon"
	"github.com/shiquda/lai/internal/logger"
	"github.com/shiquda/lai/internal/platform"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [command] [args...]",
	Short: "Monitor command output",
	Long:  "Monitor the output of a command and send notifications when threshold is reached",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var command string
		var commandArgs []string

		// If first arg contains spaces, treat it as "command args" and split it
		if len(args) == 1 && strings.Contains(args[0], " ") {
			parts := strings.Fields(args[0])
			command = parts[0]
			commandArgs = parts[1:]
		} else {
			command = args[0]
			commandArgs = args[1:]
		}

		// Get command line parameters
		lineThreshold, _ := cmd.Flags().GetInt("line-threshold")
		intervalStr, _ := cmd.Flags().GetString("interval")
		daemonMode, _ := cmd.Flags().GetBool("daemon")
		processName, _ := cmd.Flags().GetString("name")
		workingDir, _ := cmd.Flags().GetString("workdir")
		errorOnlyMode, _ := cmd.Flags().GetBool("error-only")
		finalSummary, _ := cmd.Flags().GetBool("final-summary")
		noFinalSummary, _ := cmd.Flags().GetBool("no-final-summary")
		finalSummaryOnly, _ := cmd.Flags().GetBool("final-summary-only")
		enabledNotifiers, _ := cmd.Flags().GetStringSlice("notifiers")

		// Parse time interval
		var checkInterval *time.Duration
		if intervalStr != "" {
			if duration, err := time.ParseDuration(intervalStr); err != nil {
				logger.Fatalf("Invalid interval format: %v", err)
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

		// Handle final-summary parameter
		var finalSummaryPtr *bool
		if cmd.Flags().Changed("final-summary") {
			finalSummaryPtr = &finalSummary
		} else if cmd.Flags().Changed("no-final-summary") {
			noFinalSummaryValue := !noFinalSummary
			finalSummaryPtr = &noFinalSummaryValue
		}

		// Handle final-summary-only parameter
		var finalSummaryOnlyPtr *bool
		if cmd.Flags().Changed("final-summary-only") {
			finalSummaryOnlyPtr = &finalSummaryOnly
		}

		// Logic: if final-summary-only is enabled, also enable final-summary
		if finalSummaryOnly {
			finalSummary = true
			finalSummaryPtr = &finalSummary
		}

		if daemonMode {
			if err := runStreamDaemon(command, commandArgs, lineThresholdPtr, checkInterval, processName, workingDir, finalSummaryPtr, errorOnlyModePtr, finalSummaryOnlyPtr, enabledNotifiers); err != nil {
				logger.Fatalf("Stream daemon startup failed: %v", err)
			}
		} else {
			if err := runStreamMonitor(command, commandArgs, lineThresholdPtr, checkInterval, workingDir, finalSummaryPtr, errorOnlyModePtr, finalSummaryOnlyPtr, enabledNotifiers); err != nil {
				logger.Fatalf("Stream monitor failed: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)

	// Add command line parameters (same as start command)
	execCmd.Flags().IntP("line-threshold", "l", 0, "Number of lines to trigger summary (overrides global config)")
	execCmd.Flags().StringP("interval", "i", "", "Check interval (e.g., 30s, 1m) (overrides global config)")
	execCmd.Flags().BoolP("daemon", "d", false, "Run in daemon mode (background)")
	execCmd.Flags().StringP("name", "n", "", "Custom process name (used in daemon mode)")
	execCmd.Flags().StringP("workdir", "w", "", "Working directory for command execution")
	execCmd.Flags().Bool("final-summary", false, "Enable final summary on program exit (overrides global config)")
	execCmd.Flags().Bool("no-final-summary", false, "Disable final summary on program exit")
	execCmd.Flags().BoolP("error-only", "E", false, "Only send notifications for errors and exceptions")
	execCmd.Flags().BoolP("final-summary-only", "F", false, "Only send notifications for final summary")
	execCmd.Flags().StringSlice("notifiers", []string{}, "Enable specific notifiers (comma-separated: telegram,email)")
}

func runStreamMonitor(command string, commandArgs []string, lineThreshold *int, checkInterval *time.Duration, workingDir string, finalSummary *bool, errorOnlyMode *bool, finalSummaryOnly *bool, enabledNotifiers []string) error {
	// Create command monitor source
	monitorSource := collector.NewCommandSource(command, commandArgs, workingDir)

	// Build unified configuration
	cfg, err := collector.BuildMonitorConfig(monitorSource, lineThreshold, checkInterval, nil, workingDir, finalSummary, errorOnlyMode, finalSummaryOnly)
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

func runStreamDaemon(command string, commandArgs []string, lineThreshold *int, checkInterval *time.Duration, processName string, workingDir string, finalSummary *bool, errorOnlyMode *bool, finalSummaryOnly *bool, enabledNotifiers []string) error {
	// Create daemon manager
	manager, err := daemon.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create daemon manager: %w", err)
	}

	// Generate process ID
	var processID string
	commandStr := fmt.Sprintf("%s %v", command, commandArgs)
	if processName != "" {
		processID = manager.GenerateProcessIDWithName(processName)
		// Check if process name already exists
		if manager.ProcessExists(processID) {
			return fmt.Errorf("process with name '%s' already exists", processName)
		}
	} else {
		processID = manager.GenerateProcessID(commandStr)
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

		// Save process information
		processInfo := &daemon.ProcessInfo{
			ID:        processID,
			PID:       process.Pid,
			LogFile:   commandStr,
			StartTime: time.Now(),
			Status:    "running",
		}

		if err := manager.SaveProcessInfo(processInfo); err != nil {
			return fmt.Errorf("failed to save process info: %w", err)
		}

		logger.Printf("Started stream daemon with process ID: %s (PID: %d)\n", processID, process.Pid)
		logger.Printf("Log file: %s\n", daemonLogPath)
		logger.Printf("Use 'lai list' to see running processes\n")
		logger.Printf("Use 'lai logs %s' to view logs\n", processID)
		logger.Printf("Use 'lai stop %s' to stop the process\n", processID)

		return nil
	}

	// Child process - run as daemon
	// Setup cleanup for daemon process
	defer func() {
		// Update process status to stopped instead of removing
		info, err := manager.LoadProcessInfo(processID)
		if err == nil {
			info.Status = "stopped"
			if saveErr := manager.SaveProcessInfo(info); saveErr != nil {
				logger.Errorf("Failed to update process status: %v", saveErr)
			}
		}
	}()

	// Run the actual stream monitoring
	return runStreamMonitor(command, commandArgs, lineThreshold, checkInterval, workingDir, finalSummary, errorOnlyMode, finalSummaryOnly, enabledNotifiers)
}
