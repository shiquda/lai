package cmd

import (
	"fmt"

	"github.com/shiquda/lai/internal/daemon"
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
			fmt.Printf("Failed to create daemon manager: %v\n", err)
			return
		}

		if cleanAll {
			if err := cleanAllStoppedProcesses(manager); err != nil {
				fmt.Printf("Failed to clean all stopped processes: %v\n", err)
			}
		} else if len(args) == 1 {
			processID := args[0]
			if err := cleanSingleProcess(manager, processID); err != nil {
				fmt.Printf("Failed to clean process %s: %v\n", processID, err)
			}
		} else {
			fmt.Println("Please specify a process ID or use --all flag")
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

	fmt.Printf("Cleaned up process: %s\n", processID)
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
				fmt.Printf("Failed to clean process %s: %v\n", proc.ID, err)
				continue
			}
			fmt.Printf("Cleaned up process: %s\n", proc.ID)
			cleanedCount++
		}
	}

	if cleanedCount == 0 {
		fmt.Println("No stopped processes to clean")
	} else {
		fmt.Printf("Cleaned up %d stopped processes\n", cleanedCount)
	}

	return nil
}