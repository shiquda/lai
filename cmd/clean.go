package cmd

import (
	"fmt"

	"github.com/shiquda/lai/internal/daemon"
	"github.com/shiquda/lai/internal/logger"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean [process-id]",
	Short: "Clean up stopped daemon processes",
	Long:  "Remove stopped daemon processes from the process list. Use --all to clean all stopped processes.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cleanAll, _ := cmd.Flags().GetBool("all")

		manager, err := daemon.NewManager()
		if err != nil {
			logger.Errorf("Failed to create daemon manager: %v", err)
			return
		}

		if cleanAll {
			if err := cleanAllStoppedProcesses(manager); err != nil {
				logger.Errorf("Failed to clean all stopped processes: %v", err)
			}
		} else if len(args) == 1 {
			processID := args[0]
			if err := cleanSingleProcess(manager, processID); err != nil {
				logger.Errorf("Failed to clean process %s: %v", processID, err)
			}
		} else {
			logger.UserError("Please specify a process ID or use --all flag")
		}
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().BoolP("all", "a", false, "Clean all stopped processes")
}

func cleanSingleProcess(manager *daemon.Manager, processID string) error {
	// Check if process exists
	info, err := manager.LoadProcessInfo(processID)
	if err != nil {
		return fmt.Errorf("process not found: %s", processID)
	}

	// Check if process is stopped
	if manager.IsProcessRunning(info.PID) {
		return fmt.Errorf("process %s is still running, stop it first", processID)
	}

	// Remove process info
	if err := manager.RemoveProcessInfo(processID); err != nil {
		return fmt.Errorf("failed to remove process info: %w", err)
	}

	logger.UserSuccessf("Cleaned up process: %s\n", processID)
	return nil
}

func cleanAllStoppedProcesses(manager *daemon.Manager) error {
	processes, err := manager.ListProcesses()
	if err != nil {
		return fmt.Errorf("failed to list processes: %w", err)
	}

	cleanedCount := 0
	for _, proc := range processes {
		if proc.Status == "stopped" {
			if err := manager.RemoveProcessInfo(proc.ID); err != nil {
				logger.Errorf("Failed to clean process %s: %v", proc.ID, err)
				continue
			}
			logger.UserSuccessf("Cleaned up process: %s\n", proc.ID)
			cleanedCount++
		}
	}

	if cleanedCount == 0 {
		logger.UserInfo("No stopped processes to clean")
	} else {
		logger.UserSuccessf("Cleaned up %d stopped processes\n", cleanedCount)
	}

	return nil
}
