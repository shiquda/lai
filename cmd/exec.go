package cmd

import (
	"fmt"
	"github.com/shiquda/lai/internal/collector"
	"github.com/shiquda/lai/internal/logger"
	"github.com/spf13/cobra"
)

// ExecCommandRunner command monitoring executor
type ExecCommandRunner struct {
	BaseCommandRunner
}

// NewExecCommand creates command monitoring command
func NewExecCommand() *cobra.Command {
	runner := &ExecCommandRunner{}

	cmd := &cobra.Command{
		Use:   "exec [command] [args...]",
		Short: "Monitor command output",
		Long:  "Monitor the output of a command and send notifications when threshold is reached",
		Args:  cobra.MinimumNArgs(1),
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
					logger.Fatalf("Command monitor failed: %v", err)
				}
			}
		},
	}

	// Add common parameters
	AddCommonFlags(cmd)

	return cmd
}

// ParseArgs parses exec command parameters
func (r *ExecCommandRunner) ParseArgs(cmd *cobra.Command, args []string) (*CommandOptions, collector.MonitorSource, error) {
	// Parse common parameters
	options, err := r.BaseCommandRunner.ParseCommonArgs(cmd)
	if err != nil {
		return nil, nil, err
	}

	// Parse command
	var commandStr string
	if len(args) == 1 {
		command, commandArgs, err := ParseCommandWrapper(args[0])
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse command: %w", err)
		}
		commandStr = command
		for _, arg := range commandArgs {
			commandStr += " " + arg
		}
	} else {
		commandStr = args[0]
		for _, arg := range args[1:] {
			commandStr += " " + arg
		}
	}

	// Create command monitoring source
	source := collector.NewFileSource(fmt.Sprintf("COMMAND_SOURCE:%s", commandStr))

	return options, source, nil
}