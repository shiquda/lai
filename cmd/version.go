package cmd

import (
	"runtime"

	"github.com/shiquda/lai/internal/logger"
	"github.com/shiquda/lai/internal/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the version number, build time, and git commit information for lai.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debugf("Version command called with args: %v", args)
		logger.UserInfof("lai version %s\n", version.Version)
		logger.UserInfof("Build time: %s\n", version.BuildTime)
		logger.UserInfof("Git commit: %s\n", version.GitCommit)
		logger.UserInfof("Go version: %s\n", runtime.Version())
		logger.UserInfof("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		logger.Debugf("Version information displayed successfully")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
