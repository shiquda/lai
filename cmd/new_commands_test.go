package cmd

import (
	"testing"
	"time"
)

func TestNewFileCommand(t *testing.T) {
	cmd := NewFileCommand()

	if cmd == nil {
		t.Fatal("NewFileCommand returned nil")
	}

	if cmd.Use != "file [log-file]" {
		t.Errorf("Expected command use 'file [log-file]', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command short description is empty")
	}

	// Check that the command validates arguments correctly
	if cmd.Args == nil {
		t.Error("Command args validation is not set")
	}

	// Check that required flags are present
	flags := []string{"line-threshold", "interval", "chat-id", "name", "workdir", "final-summary", "error-only", "final-summary-only", "notifiers", "daemon"}
	for _, flag := range flags {
		if f := cmd.Flag(flag); f == nil {
			t.Errorf("Flag '%s' not found", flag)
		}
	}
}

func TestNewExecCommand(t *testing.T) {
	cmd := NewExecCommand()

	if cmd == nil {
		t.Fatal("NewExecCommand returned nil")
	}

	if cmd.Use != "exec [command] [args...]" {
		t.Errorf("Expected command use 'exec [command] [args...]', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Command short description is empty")
	}

	// Check that required flags are present
	flags := []string{"line-threshold", "interval", "chat-id", "name", "workdir", "final-summary", "error-only", "final-summary-only", "notifiers", "daemon"}
	for _, flag := range flags {
		if f := cmd.Flag(flag); f == nil {
			t.Errorf("Flag '%s' not found", flag)
		}
	}
}

func TestCommandOptions(t *testing.T) {
	// Test CommandOptions struct creation
	opts := &CommandOptions{
		LineThreshold:    func() *int { i := 10; return &i }(),
		CheckInterval:    func() *time.Duration { d := 30 * time.Second; return &d }(),
		ChatID:           func() *string { s := "test-chat"; return &s }(),
		ProcessName:      "test-process",
		WorkingDir:       "/tmp",
		FinalSummary:     func() *bool { b := true; return &b }(),
		ErrorOnlyMode:    func() *bool { b := false; return &b }(),
		FinalSummaryOnly: func() *bool { b := false; return &b }(),
		EnabledNotifiers: []string{"telegram", "email"},
		DaemonMode:       true,
	}

	// Test that values are set correctly
	if opts.ProcessName != "test-process" {
		t.Errorf("Expected ProcessName 'test-process', got '%s'", opts.ProcessName)
	}

	if opts.WorkingDir != "/tmp" {
		t.Errorf("Expected WorkingDir '/tmp', got '%s'", opts.WorkingDir)
	}

	if !opts.DaemonMode {
		t.Error("Expected DaemonMode to be true")
	}

	if len(opts.EnabledNotifiers) != 2 {
		t.Errorf("Expected 2 enabled notifiers, got %d", len(opts.EnabledNotifiers))
	}
}

func TestParseCommandWrapperIntegration(t *testing.T) {
	// Test that ParseCommandWrapper works with the new commands
	testCases := []struct {
		name        string
		input       string
		expectCmd   string
		expectArgs  []string
		expectError bool
	}{
		{
			name:        "Simple command",
			input:       "ls -la",
			expectCmd:   "ls",
			expectArgs:  []string{"-la"},
			expectError: false,
		},
		{
			name:        "Command with quotes",
			input:       `echo "hello world"`,
			expectCmd:   "echo",
			expectArgs:  []string{"hello world"},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, args, err := ParseCommandWrapper(tc.input)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if cmd != tc.expectCmd {
				t.Errorf("Expected command '%s', got '%s'", tc.expectCmd, cmd)
			}

			if len(args) != len(tc.expectArgs) {
				t.Errorf("Expected %d args, got %d", len(tc.expectArgs), len(args))
				return
			}

			for i, expectedArg := range tc.expectArgs {
				if args[i] != expectedArg {
					t.Errorf("Expected arg[%d] '%s', got '%s'", i, expectedArg, args[i])
				}
			}
		})
	}
}
