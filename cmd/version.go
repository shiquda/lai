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
		logger.UserInfof("lai version %s", version.Version)
		logger.UserInfof("Build time: %s", version.BuildTime)
		logger.UserInfof("Git commit: %s", version.GitCommit)
		logger.UserInfof("Go version: %s", runtime.Version())
		logger.UserInfof("OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
		logger.Debugf("Version information displayed successfully")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
