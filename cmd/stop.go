package cmd

import (
	"github.com/shiquda/lai/internal/daemon"
	"github.com/shiquda/lai/internal/logger"
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
			logger.Errorf("Failed to create daemon manager: %v", err)
			return
		}

		if all {
			logger.UserInfo("Stopping all daemon processes...")
			if err := manager.StopAllProcesses(); err != nil {
				logger.Errorf("Failed to stop all processes: %v", err)
				return
			}
			logger.UserSuccess("All processes stopped successfully")
			return
		}

		if len(args) == 0 {
			logger.UserError("Error: process-id is required (or use --all flag)")
			cmd.Usage()
			return
		}

		processID := args[0]

		// Check if process exists
		processInfo, err := manager.LoadProcessInfo(processID)
		if err != nil {
			logger.Errorf("Process not found: %s", processID)
			return
		}

		logger.UserInfof("Stopping process: %s (PID: %d)", processID, processInfo.PID)

		if err := manager.StopProcess(processID); err != nil {
			logger.Errorf("Failed to stop process: %v", err)
			return
		}

		logger.UserSuccessf("Process %s stopped successfully", processID)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().BoolP("all", "a", false, "Stop all daemon processes")
}
