package cmd

import (
	"bufio"
	"os"

	"github.com/shiquda/lai/internal/daemon"
	"github.com/shiquda/lai/internal/logger"
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
			logger.Errorf("Failed to create daemon manager: %v", err)
			return
		}

		// Check if process exists
		_, err = manager.LoadProcessInfo(processID)
		if err != nil {
			logger.Errorf("Process not found: %s", processID)
			return
		}

		logPath := manager.GetProcessLogPath(processID)

		// Check if log file exists
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			logger.Errorf("Log file not found: %s", logPath)
			return
		}

		if follow {
			if err := tailFile(logPath); err != nil {
				logger.Errorf("Failed to tail log file: %v", err)
			}
		} else {
			if err := showLastLines(logPath, lines); err != nil {
				logger.Errorf("Failed to show log file: %v", err)
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
		logger.Println(lines[i])
	}

	return nil
}

func tailFile(filePath string) error {
	// Simple tail implementation - show existing content and then follow
	if err := showLastLines(filePath, 50); err != nil {
		return err
	}

	logger.Println("==> Following log file (Ctrl+C to stop) <==")

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
		logger.Println(scanner.Text())
	}

	return scanner.Err()
}
