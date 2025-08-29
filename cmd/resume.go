package cmd

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/shiquda/lai/internal/daemon"
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
			fmt.Printf("Failed to create daemon manager: %v\n", err)
			return
		}

		if err := resumeProcess(manager, processID); err != nil {
			log.Fatalf("Failed to resume process %s: %v", processID, err)
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

	// Temporarily remove the existing process info to avoid duplicate name error
	originalLogFile := info.LogFile
	if err := manager.RemoveProcessInfo(processID); err != nil {
		return fmt.Errorf("failed to remove existing process info: %w", err)
	}

	// Get log file path for daemon
	daemonLogPath := manager.GetProcessLogPath(processID)

	// Set environment variable for daemon mode
	os.Setenv("LAI_DAEMON_MODE", "1")

	// Redirect output to log file
	logFileHandle, err := os.OpenFile(daemonLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
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

	procAttr := &os.ProcAttr{
		Files: []*os.File{nil, logFileHandle, logFileHandle},
		Env:   append(os.Environ(), "LAI_DAEMON_MODE=1"),
		Sys:   &syscall.SysProcAttr{Setsid: true},
	}

	process, err := os.StartProcess(cmd, args, procAttr)
	if err != nil {
		// If failed to start, restore the original process info
		info.Status = "stopped"
		manager.SaveProcessInfo(info)
		return fmt.Errorf("failed to start daemon process: %w", err)
	}

	fmt.Printf("Resumed daemon with process ID: %s (PID: %d)\n", processID, process.Pid)
	fmt.Printf("Log file: %s\n", daemonLogPath)
	fmt.Printf("Use 'lai list' to see running processes\n")
	fmt.Printf("Use 'lai logs %s' to view logs\n", processID)
	fmt.Printf("Use 'lai stop %s' to stop the process\n", processID)

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