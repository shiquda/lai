package cmd

import (
	"github.com/shiquda/lai/internal/daemon"
	"github.com/shiquda/lai/internal/logger"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List running daemon processes",
	Long:  "List all currently running daemon processes with their status",
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := daemon.NewManager()
		if err != nil {
			logger.Errorf("Failed to create daemon manager: %v", err)
			return
		}

		processes, err := manager.ListProcesses()
		if err != nil {
			logger.Errorf("Failed to list processes: %v", err)
			return
		}

		if len(processes) == 0 {
			logger.Println("No daemon processes found")
			return
		}

		logger.Printf("%-20s %-8s %-10s %-20s %s\n", "PROCESS ID", "PID", "STATUS", "START TIME", "LOG FILE")
		logger.Printf("%-20s %-8s %-10s %-20s %s\n", "----------", "---", "------", "----------", "--------")

		for _, proc := range processes {
			startTime := proc.StartTime.Format("2006-01-02 15:04:05")
			logger.Printf("%-20s %-8d %-10s %-20s %s\n",
				proc.ID, proc.PID, proc.Status, startTime, proc.LogFile)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
