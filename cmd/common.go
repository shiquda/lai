package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/shiquda/lai/internal/collector"
	"github.com/shiquda/lai/internal/daemon"
	"github.com/shiquda/lai/internal/logger"
	"github.com/shiquda/lai/internal/platform"
	"github.com/spf13/cobra"
)

// CommandOptions contains common parameters for all commands
type CommandOptions struct {
	LineThreshold    *int
	CheckInterval    *time.Duration
	ChatID           *string
	ProcessName      string
	WorkingDir       string
	FinalSummary     *bool
	ErrorOnlyMode    *bool
	FinalSummaryOnly *bool
	EnabledNotifiers []string
	DaemonMode       bool
}

// CommandRunner defines command execution interface
type CommandRunner interface {
	ParseArgs(cmd *cobra.Command, args []string) (*CommandOptions, collector.MonitorSource, error)
	Run(options *CommandOptions, source collector.MonitorSource) error
	RunDaemon(options *CommandOptions, source collector.MonitorSource) error
}

// BaseCommandRunner base command executor
type BaseCommandRunner struct{}

// ParseCommonArgs parses common parameters
func (r *BaseCommandRunner) ParseCommonArgs(cmd *cobra.Command) (*CommandOptions, error) {
	options := &CommandOptions{}

	// Parse basic parameters
	lineThreshold, _ := cmd.Flags().GetInt("line-threshold")
	if cmd.Flags().Changed("line-threshold") {
		options.LineThreshold = &lineThreshold
	}

	intervalStr, _ := cmd.Flags().GetString("interval")
	if intervalStr != "" {
		if duration, err := time.ParseDuration(intervalStr); err != nil {
			return nil, fmt.Errorf("invalid interval format: %v", err)
		} else {
			options.CheckInterval = &duration
		}
	}

	chatID, _ := cmd.Flags().GetString("chat-id")
	if cmd.Flags().Changed("chat-id") {
		options.ChatID = &chatID
	}

	options.ProcessName, _ = cmd.Flags().GetString("name")
	options.WorkingDir, _ = cmd.Flags().GetString("workdir")

	errorOnlyMode, _ := cmd.Flags().GetBool("error-only")
	if cmd.Flags().Changed("error-only") {
		options.ErrorOnlyMode = &errorOnlyMode
	}

	finalSummary, _ := cmd.Flags().GetBool("final-summary")
	noFinalSummary, _ := cmd.Flags().GetBool("no-final-summary")
	finalSummaryOnly, _ := cmd.Flags().GetBool("final-summary-only")

	// Handle final-summary parameter
	if cmd.Flags().Changed("final-summary") {
		options.FinalSummary = &finalSummary
	} else if cmd.Flags().Changed("no-final-summary") {
		noFinalSummaryValue := !noFinalSummary
		options.FinalSummary = &noFinalSummaryValue
	}

	// Handle final-summary-only parameter
	if cmd.Flags().Changed("final-summary-only") {
		options.FinalSummaryOnly = &finalSummaryOnly
	}

	// Logic: if final-summary-only is enabled, also enable final-summary
	if finalSummaryOnly {
		finalSummary = true
		options.FinalSummary = &finalSummary
	}

	options.EnabledNotifiers, _ = cmd.Flags().GetStringSlice("notifiers")
	options.DaemonMode, _ = cmd.Flags().GetBool("daemon")

	return options, nil
}

// Run executes monitoring
func (r *BaseCommandRunner) Run(options *CommandOptions, source collector.MonitorSource) error {
	cfg, err := collector.BuildMonitorConfig(
		source,
		options.LineThreshold,
		options.CheckInterval,
		options.ChatID,
		options.WorkingDir,
		options.FinalSummary,
		options.ErrorOnlyMode,
		options.FinalSummaryOnly,
	)
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	monitor, err := collector.NewUnifiedMonitor(cfg)
	if err != nil {
		return fmt.Errorf("failed to create unified monitor: %w", err)
	}

	return monitor.Start()
}

// RunDaemon runs daemon process
func (r *BaseCommandRunner) RunDaemon(options *CommandOptions, source collector.MonitorSource) error {
	manager, err := daemon.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create daemon manager: %w", err)
	}

	sourceIdentifier := source.GetIdentifier()
	processID := r.generateProcessID(manager, options.ProcessName, sourceIdentifier)
	daemonLogPath := manager.GetProcessLogPath(processID)

	// Parent process - start daemon process
	if os.Getenv("LAI_DAEMON_MODE") != "1" {
		return r.startDaemonProcess(manager, processID, sourceIdentifier, daemonLogPath)
	}

	// Child process - run as daemon
	return r.runAsDaemon(manager, processID, options, source)
}

// generateProcessID generates process ID
func (r *BaseCommandRunner) generateProcessID(manager *daemon.Manager, processName, sourceIdentifier string) string {
	if processName != "" {
		processID := manager.GenerateProcessIDWithName(processName)
		if manager.ProcessExists(processID) {
			logger.Fatalf("Process with name '%s' already exists", processName)
		}
		return processID
	}
	return manager.GenerateProcessID(sourceIdentifier)
}

// startDaemonProcess starts daemon process
func (r *BaseCommandRunner) startDaemonProcess(manager *daemon.Manager, processID, sourceIdentifier, daemonLogPath string) error {
	os.Setenv("LAI_DAEMON_MODE", "1")

	logFileHandle, err := os.OpenFile(daemonLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer logFileHandle.Close()

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cmd := execPath
	args := append([]string{cmd}, os.Args[1:]...)

	p := platform.New()
	process, err := p.Process.StartDaemonProcess(cmd, args, logFileHandle, append(os.Environ(), "LAI_DAEMON_MODE=1"))
	if err != nil {
		return fmt.Errorf("failed to start daemon process: %w", err)
	}

	// Save process information
	processInfo := &daemon.ProcessInfo{
		ID:        processID,
		PID:       process.Pid,
		LogFile:   sourceIdentifier,
		StartTime: time.Now(),
		Status:    "running",
	}

	if err := manager.SaveProcessInfo(processInfo); err != nil {
		return fmt.Errorf("failed to save process info: %w", err)
	}

	logger.UserSuccessf("Started daemon with process ID: %s (PID: %d)\n", processID, process.Pid)
	logger.UserInfof("Log file: %s\n", daemonLogPath)
	logger.UserInfo("Use 'lai list' to see running processes")
	logger.UserInfof("Use 'lai logs %s' to view logs", processID)
	logger.UserInfof("Use 'lai stop %s' to stop the process", processID)

	return nil
}

// runAsDaemon runs as daemon process
func (r *BaseCommandRunner) runAsDaemon(manager *daemon.Manager, processID string, options *CommandOptions, source collector.MonitorSource) error {
	// Set cleanup function
	defer func() {
		if info, err := manager.LoadProcessInfo(processID); err == nil {
			info.Status = "stopped"
			if saveErr := manager.SaveProcessInfo(info); saveErr != nil {
				logger.Errorf("Failed to update process status: %v", saveErr)
			}
		}
	}()

	// Update child process information
	processInfo := &daemon.ProcessInfo{
		ID:        processID,
		PID:       os.Getpid(),
		LogFile:   source.GetIdentifier(),
		StartTime: time.Now(),
		Status:    "running",
	}
	if err := manager.SaveProcessInfo(processInfo); err != nil {
		logger.Errorf("Failed to save process info in child: %v", err)
	}

	return r.Run(options, source)
}

// AddCommonFlags adds common parameters to command
func AddCommonFlags(cmd *cobra.Command) {
	cmd.Flags().IntP("line-threshold", "l", 0, "Number of lines to trigger summary (overrides global config)")
	cmd.Flags().StringP("interval", "i", "", "Check interval (e.g., 30s, 1m) (overrides global config)")
	cmd.Flags().StringP("chat-id", "c", "", "Telegram chat ID (overrides global config)")
	cmd.Flags().BoolP("daemon", "d", false, "Run in daemon mode (background)")
	cmd.Flags().StringP("name", "n", "", "Custom process name (used in daemon mode)")
	cmd.Flags().StringP("workdir", "w", "", "Working directory for command execution")
	cmd.Flags().Bool("final-summary", false, "Enable final summary on program exit (overrides global config)")
	cmd.Flags().Bool("no-final-summary", false, "Disable final summary on program exit")
	cmd.Flags().BoolP("error-only", "E", false, "Only send notifications for errors and exceptions")
	cmd.Flags().BoolP("final-summary-only", "F", false, "Only send notifications for final summary")
	cmd.Flags().StringSlice("notifiers", []string{}, "Enable specific notifiers (comma-separated: telegram,email)")
}
