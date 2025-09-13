package cmd

import (
	"github.com/shiquda/lai/internal/collector"
	"github.com/shiquda/lai/internal/logger"
	"github.com/spf13/cobra"
)

// FileCommandRunner file monitoring command executor
type FileCommandRunner struct {
	BaseCommandRunner
}

// NewFileCommand creates file monitoring command
func NewFileCommand() *cobra.Command {
	runner := &FileCommandRunner{}

	cmd := &cobra.Command{
		Use:   "file [log-file]",
		Short: "Monitor log file",
		Long:  "Monitor log file and send notifications when new content is detected",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			options, source, err := runner.ParseArgs(cmd, args)
			if err != nil {
				logger.Fatalf("Failed to parse arguments: %v", err)
			}

			if options.DaemonMode {
				if err := runner.RunDaemon(options, source); err != nil {
					logger.Fatalf("Daemon startup failed: %v", err)
				}
			} else {
				if err := runner.Run(options, source); err != nil {
					logger.Fatalf("File monitor failed: %v", err)
				}
			}
		},
	}

	// Add common parameters
	AddCommonFlags(cmd)

	return cmd
}

// ParseArgs parses file command parameters
func (r *FileCommandRunner) ParseArgs(cmd *cobra.Command, args []string) (*CommandOptions, collector.MonitorSource, error) {
	// Parse common parameters
	options, err := r.BaseCommandRunner.ParseCommonArgs(cmd)
	if err != nil {
		return nil, nil, err
	}

	// Create file monitoring source
	logFile := args[0]
	source := collector.NewFileSource(logFile)

	return options, source, nil
}
