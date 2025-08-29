package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/shiquda/lai/internal/daemon"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs <process-id>",
	Short: "Show logs for a daemon process",
	Long:  "Show logs for the specified daemon process",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		processID := args[0]
		follow, _ := cmd.Flags().GetBool("follow")
		lines, _ := cmd.Flags().GetInt("lines")

		manager, err := daemon.NewManager()
		if err != nil {
			fmt.Printf("Failed to create daemon manager: %v\n", err)
			return
		}

		// Check if process exists
		_, err = manager.LoadProcessInfo(processID)
		if err != nil {
			fmt.Printf("Process not found: %s\n", processID)
			return
		}

		logPath := manager.GetProcessLogPath(processID)

		// Check if log file exists
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			fmt.Printf("Log file not found: %s\n", logPath)
			return
		}

		if follow {
			if err := tailFile(logPath); err != nil {
				fmt.Printf("Failed to tail log file: %v\n", err)
			}
		} else {
			if err := showLastLines(logPath, lines); err != nil {
				fmt.Printf("Failed to show log file: %v\n", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	logsCmd.Flags().IntP("lines", "n", 50, "Number of lines to show from the end")
}

func showLastLines(filePath string, numLines int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read all lines into a slice
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Show last N lines
	start := len(lines) - numLines
	if start < 0 {
		start = 0
	}

	for i := start; i < len(lines); i++ {
		fmt.Println(lines[i])
	}

	return nil
}

func tailFile(filePath string) error {
	// Simple tail implementation - show existing content and then follow
	if err := showLastLines(filePath, 50); err != nil {
		return err
	}

	fmt.Println("==> Following log file (Ctrl+C to stop) <==")

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Seek to end of file
	if _, err := file.Seek(0, 2); err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return scanner.Err()
}
