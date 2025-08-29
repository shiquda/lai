package cmd

import (
	"fmt"

	"github.com/shiquda/lai/internal/daemon"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop <process-id>",
	Short: "Stop a daemon process",
	Long:  "Stop the specified daemon process or all processes with --all flag",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")

		manager, err := daemon.NewManager()
		if err != nil {
			fmt.Printf("Failed to create daemon manager: %v\n", err)
			return
		}

		if all {
			fmt.Println("Stopping all daemon processes...")
			if err := manager.StopAllProcesses(); err != nil {
				fmt.Printf("Failed to stop all processes: %v\n", err)
				return
			}
			fmt.Println("All processes stopped successfully")
			return
		}

		if len(args) == 0 {
			fmt.Println("Error: process-id is required (or use --all flag)")
			cmd.Usage()
			return
		}

		processID := args[0]

		// Check if process exists
		processInfo, err := manager.LoadProcessInfo(processID)
		if err != nil {
			fmt.Printf("Process not found: %s\n", processID)
			return
		}

		fmt.Printf("Stopping process: %s (PID: %d)\n", processID, processInfo.PID)

		if err := manager.StopProcess(processID); err != nil {
			fmt.Printf("Failed to stop process: %v\n", err)
			return
		}

		fmt.Printf("Process %s stopped successfully\n", processID)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().BoolP("all", "a", false, "Stop all daemon processes")
}