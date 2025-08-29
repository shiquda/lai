package cmd

import (
	"fmt"

	"github.com/shiquda/lai/internal/daemon"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List running daemon processes",
	Long:  "List all currently running daemon processes with their status",
	Run: func(cmd *cobra.Command, args []string) {
		manager, err := daemon.NewManager()
		if err != nil {
			fmt.Printf("Failed to create daemon manager: %v\n", err)
			return
		}

		processes, err := manager.ListProcesses()
		if err != nil {
			fmt.Printf("Failed to list processes: %v\n", err)
			return
		}

		if len(processes) == 0 {
			fmt.Println("No daemon processes found")
			return
		}

		fmt.Printf("%-20s %-8s %-10s %-20s %s\n", "PROCESS ID", "PID", "STATUS", "START TIME", "LOG FILE")
		fmt.Printf("%-20s %-8s %-10s %-20s %s\n", "----------", "---", "------", "----------", "--------")

		for _, proc := range processes {
			startTime := proc.StartTime.Format("2006-01-02 15:04:05")
			fmt.Printf("%-20s %-8d %-10s %-20s %s\n",
				proc.ID, proc.PID, proc.Status, startTime, proc.LogFile)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
