package cmd

import (
	"os"

	"github.com/shiquda/lai/internal/config"
	"github.com/shiquda/lai/internal/logger"
	"github.com/shiquda/lai/internal/version"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "lai",
	Short:   "A smart lightweight log monitoring and notification tool",
	Long:    `A smart lightweight log monitoring and notification tool`,
	Version: version.Version,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Initialize logging system
	if err := config.InitLogger(); err != nil {
		logger.Fatalf("Failed to initialize logging system: %v", err)
	}

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Register new commands
	rootCmd.AddCommand(NewFileCommand())
	rootCmd.AddCommand(NewExecCommand())
}
