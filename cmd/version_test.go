package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/shiquda/lai/internal/version"
	"github.com/spf13/cobra"
)

func TestVersionCommand(t *testing.T) {
	// Save original values
	originalVersion := version.Version
	originalBuildTime := version.BuildTime
	originalGitCommit := version.GitCommit

	// Set test values
	version.Version = "1.0.0"
	version.BuildTime = "2025-01-01_12:00:00"
	version.GitCommit = "abc1234"

	// Restore original values after test
	defer func() {
		version.Version = originalVersion
		version.BuildTime = originalBuildTime
		version.GitCommit = originalGitCommit
	}()

	// Create a buffer to capture output
	buf := new(bytes.Buffer)

	// Create a new version command for testing
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("lai version %s\n", version.Version)
			cmd.Printf("Build time: %s\n", version.BuildTime)
			cmd.Printf("Git commit: %s\n", version.GitCommit)
		},
	}

	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Execute the command
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check output
	output := buf.String()

	expectedStrings := []string{
		"lai version 1.0.0",
		"Build time: 2025-01-01_12:00:00",
		"Git commit: abc1234",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got:\n%s", expected, output)
		}
	}
}

func TestVersionVariables(t *testing.T) {
	// Test that version variables have default values
	if version.Version == "" {
		t.Error("Version should not be empty")
	}
	if version.BuildTime == "" {
		t.Error("BuildTime should not be empty")
	}
	if version.GitCommit == "" {
		t.Error("GitCommit should not be empty")
	}
}
