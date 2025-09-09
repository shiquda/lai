package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/shiquda/lai/internal/collector"
	"github.com/shiquda/lai/internal/config"
	"github.com/shiquda/lai/internal/daemon"
	"github.com/shiquda/lai/internal/notifier"
	"github.com/shiquda/lai/internal/platform"
	"github.com/shiquda/lai/internal/summarizer"
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
		chatID, _ := cmd.Flags().GetString("chat-id")
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

		// Handle chat-id parameter
		var chatIDPtr *string
		if cmd.Flags().Changed("chat-id") {
			chatIDPtr = &chatID
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
			if err := runStreamDaemon(command, commandArgs, lineThresholdPtr, checkInterval, chatIDPtr, processName, workingDir, finalSummaryPtr, errorOnlyModePtr, finalSummaryOnlyPtr, enabledNotifiers); err != nil {
				log.Fatalf("Stream daemon startup failed: %v", err)
			}
		} else {
			if err := runStreamMonitor(command, commandArgs, lineThresholdPtr, checkInterval, chatIDPtr, workingDir, finalSummaryPtr, errorOnlyModePtr, finalSummaryOnlyPtr, enabledNotifiers); err != nil {
				log.Fatalf("Stream monitor failed: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(execCmd)

	// Add command line parameters (same as start command)
	execCmd.Flags().IntP("line-threshold", "l", 0, "Number of lines to trigger summary (overrides global config)")
	execCmd.Flags().StringP("interval", "i", "", "Check interval (e.g., 30s, 1m) (overrides global config)")
	execCmd.Flags().StringP("chat-id", "c", "", "Telegram chat ID (overrides global config)")
	execCmd.Flags().BoolP("daemon", "d", false, "Run in daemon mode (background)")
	execCmd.Flags().StringP("name", "n", "", "Custom process name (used in daemon mode)")
	execCmd.Flags().StringP("workdir", "w", "", "Working directory for command execution")
	execCmd.Flags().Bool("final-summary", false, "Enable final summary on program exit (overrides global config)")
	execCmd.Flags().Bool("no-final-summary", false, "Disable final summary on program exit")
	execCmd.Flags().BoolP("error-only", "E", false, "Only send notifications for errors and exceptions")
	execCmd.Flags().BoolP("final-summary-only", "F", false, "Only send notifications for final summary")
	execCmd.Flags().StringSlice("notifiers", []string{}, "Enable specific notifiers (comma-separated: telegram,email)")
}

func runStreamMonitor(command string, commandArgs []string, lineThreshold *int, checkInterval *time.Duration, chatID *string, workingDir string, finalSummary *bool, errorOnlyMode *bool, finalSummaryOnly *bool, enabledNotifiers []string) error {
	// Build runtime configuration for stream monitoring
	cfg, err := config.BuildStreamConfig(command, commandArgs, lineThreshold, checkInterval, chatID, workingDir, finalSummary, errorOnlyMode, finalSummaryOnly)
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Create service instances
	openaiClient := summarizer.NewOpenAIClient(cfg.OpenAI.APIKey, cfg.OpenAI.BaseURL, cfg.OpenAI.Model)
	
	// Create notifiers using factory function (same as start command)
	notifiers, err := notifier.CreateNotifiers(cfg, enabledNotifiers)
	if err != nil {
		return fmt.Errorf("failed to create notifiers: %w", err)
	}
	
	streamCollector := collector.NewStreamCollector(cfg.Command, cfg.CommandArgs, cfg.LineThreshold, cfg.CheckInterval, cfg.FinalSummary)

	// Set trigger handler
	streamCollector.SetTriggerHandler(func(newContent string) error {
		// If FinalSummaryOnly mode is enabled, skip intermediate notifications
		if cfg.FinalSummaryOnly {
			fmt.Printf("Command output changes detected, but skipping intermediate notification (final-summary-only mode)\n")
			return nil
		}

		fmt.Printf("Command output changes detected, generating summary...\n")

		summary, err := openaiClient.Summarize(newContent, cfg.Language)
		if err != nil {
			return fmt.Errorf("failed to generate summary: %w", err)
		}

		commandStr := fmt.Sprintf("%s %v", cfg.Command, cfg.CommandArgs)
		if err := sendToAllNotifiersForExec(notifiers, commandStr, summary); err != nil {
			return fmt.Errorf("failed to send notification: %w", err)
		}

		fmt.Printf("Summary sent to %d notifier(s): %v\n", len(notifiers), enabledNotifiers)
		return nil
	})

	// Display startup information
	fmt.Printf("Starting command monitoring: %s %v\n", cfg.Command, cfg.CommandArgs)
	fmt.Printf("Line threshold: %d lines\n", cfg.LineThreshold)
	fmt.Printf("Check interval: %v\n", cfg.CheckInterval)
	fmt.Printf("Chat ID: %s\n", cfg.ChatID)
	fmt.Printf("Enabled notifiers: %v\n", enabledNotifiers)
	if cfg.WorkingDir != "" {
		fmt.Printf("Working directory: %s\n", cfg.WorkingDir)
	}

	// Setup signal handling
	p := platform.New()
	sigChan := p.Signal.SetupShutdownSignals()

	// Run collector in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- streamCollector.Start()
	}()

	// Wait for signal or error
	select {
	case <-sigChan:
		fmt.Println("\nReceived stop signal, shutting down...")
		streamCollector.Stop()
		return nil
	case err := <-errChan:
		return err
	}
}

func runStreamDaemon(command string, commandArgs []string, lineThreshold *int, checkInterval *time.Duration, chatID *string, processName string, workingDir string, finalSummary *bool, errorOnlyMode *bool, finalSummaryOnly *bool, enabledNotifiers []string) error {
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

		fmt.Printf("Started stream daemon with process ID: %s (PID: %d)\n", processID, process.Pid)
		fmt.Printf("Log file: %s\n", daemonLogPath)
		fmt.Printf("Use 'lai list' to see running processes\n")
		fmt.Printf("Use 'lai logs %s' to view logs\n", processID)
		fmt.Printf("Use 'lai stop %s' to stop the process\n", processID)

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
				log.Printf("Failed to update process status: %v", saveErr)
			}
		}
	}()

	// Run the actual stream monitoring
	return runStreamMonitor(command, commandArgs, lineThreshold, checkInterval, chatID, workingDir, finalSummary, errorOnlyMode, finalSummaryOnly, enabledNotifiers)
}

// sendToAllNotifiersForExec sends a log summary to all configured notifiers for exec command
func sendToAllNotifiersForExec(notifiers []notifier.Notifier, commandStr, summary string) error {
	var errors []error
	
	for _, n := range notifiers {
		if err := n.SendLogSummary(commandStr, summary); err != nil {
			errors = append(errors, err)
			log.Printf("Failed to send notification to notifier: %v", err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("%d notifier(s) failed: %v", len(errors), errors)
	}
	
	return nil
}
