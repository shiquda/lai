package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/shiquda/lai/internal/daemon"
	"github.com/shiquda/lai/internal/logger"
	"github.com/shiquda/lai/internal/platform"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume <process-id>",
	Short: "Resume a stopped daemon process",
	Long:  "Restart a stopped daemon process with the same configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		processID := args[0]

		manager, err := daemon.NewManager()
		if err != nil {
			logger.Errorf("Failed to create daemon manager: %v", err)
			return
		}

		if err := resumeProcess(manager, processID); err != nil {
			logger.Fatalf("Failed to resume process %s: %v", processID, err)
		}
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}

func resumeProcess(manager *daemon.Manager, processID string) error {
	// Load existing process info
	info, err := manager.LoadProcessInfo(processID)
	if err != nil {
		return fmt.Errorf("process not found: %s", processID)
	}

	// Check if process is already running
	if manager.IsProcessRunning(info.PID) {
		return fmt.Errorf("process %s is already running", processID)
	}

	// Save original log file, but keep process info intact
	originalLogFile := info.LogFile

	// Mark process as resuming to prevent duplicate operations
	info.Status = "resuming"
	if err := manager.SaveProcessInfo(info); err != nil {
		return fmt.Errorf("failed to update process status: %w", err)
	}

	// Get log file path for daemon
	daemonLogPath := manager.GetProcessLogPath(processID)

	// Set environment variable for daemon mode
	os.Setenv("LAI_DAEMON_MODE", "1")

	// Redirect output to log file
	logFileHandle, err := os.OpenFile(daemonLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Restore status on error
		info.Status = "stopped"
		manager.SaveProcessInfo(info)
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer logFileHandle.Close()

	// Start daemon process
	cmd := os.Args[0]
	args := []string{cmd, "start", originalLogFile, "-d"}

	// Add the original process name if it was a custom name (no timestamp)
	if !containsTimestamp(processID) {
		args = append(args, "-n", processID)
	}

	// Use platform-specific daemon process creation
	p := platform.New()
	process, err := p.Process.StartDaemonProcess(cmd, args, logFileHandle, append(os.Environ(), "LAI_DAEMON_MODE=1", "LAI_RESUME_MODE=1"))
	if err != nil {
		// Restore status on error
		info.Status = "stopped"
		manager.SaveProcessInfo(info)
		return fmt.Errorf("failed to start daemon process: %w", err)
	}

	// Process started successfully, update with new PID and status
	info.PID = process.Pid
	info.Status = "running"
	info.StartTime = time.Now()
	if err := manager.SaveProcessInfo(info); err != nil {
		logger.Warnf("Warning: failed to update process info: %v", err)
	}

	logger.UserSuccessf("Resumed daemon with process ID: %s (PID: %d)\n", processID, process.Pid)
	logger.UserInfof("Log file: %s\n", daemonLogPath)
	logger.UserInfo("Use 'lai list' to see running processes")
	logger.UserInfof("Use 'lai logs %s' to view logs", processID)
	logger.UserInfof("Use 'lai stop %s' to stop the process", processID)

	return nil
}

// Helper function to check if process ID contains timestamp
func containsTimestamp(processID string) bool {
	// Simple heuristic: if it contains underscore followed by digits, likely has timestamp
	for i := 0; i < len(processID)-1; i++ {
		if processID[i] == '_' {
			// Check if followed by digits
			hasDigits := false
			for j := i + 1; j < len(processID); j++ {
				if processID[j] >= '0' && processID[j] <= '9' {
					hasDigits = true
				} else {
					break
				}
			}
			if hasDigits {
				return true
			}
		}
	}
	return false
}
