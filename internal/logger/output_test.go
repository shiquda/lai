package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestConsoleOutput(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer
	output := NewConsoleOutputWithWriter(&buf)

	// Test Info methods
	output.Info("info message")
	if !strings.Contains(buf.String(), "info message") {
		t.Errorf("Info() failed to output message")
	}
	buf.Reset()

	output.Infof("info %s", "formatted")
	if !strings.Contains(buf.String(), "info formatted") {
		t.Errorf("Infof() failed to format and output message")
	}
	buf.Reset()

	// Test Success methods
	output.Success("success message")
	if !strings.Contains(buf.String(), "success message") {
		t.Errorf("Success() failed to output message")
	}
	buf.Reset()

	output.Successf("success %s", "formatted")
	if !strings.Contains(buf.String(), "success formatted") {
		t.Errorf("Successf() failed to format and output message")
	}
	buf.Reset()

	// Test Warning methods
	output.Warning("warning message")
	if !strings.Contains(buf.String(), "warning message") {
		t.Errorf("Warning() failed to output message")
	}
	buf.Reset()

	output.Warningf("warning %s", "formatted")
	if !strings.Contains(buf.String(), "warning formatted") {
		t.Errorf("Warningf() failed to format and output message")
	}
	buf.Reset()

	// Test Error methods
	output.Error("error message")
	if !strings.Contains(buf.String(), "error message") {
		t.Errorf("Error() failed to output message")
	}
	buf.Reset()

	output.Errorf("error %s", "formatted")
	if !strings.Contains(buf.String(), "error formatted") {
		t.Errorf("Errorf() failed to format and output message")
	}
	buf.Reset()

}

func TestGlobalUserOutputFunctions(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer
	customOutput := NewConsoleOutputWithWriter(&buf)

	// Set custom output for testing
	oldOutput := defaultUserOutput
	defer func() {
		defaultUserOutput = oldOutput
	}()
	SetDefaultUserOutput(customOutput)

	// Test global functions
	UserInfo("global info")
	if !strings.Contains(buf.String(), "global info") {
		t.Errorf("UserInfo() failed to output message")
	}
	buf.Reset()

	UserInfof("global %s", "formatted")
	if !strings.Contains(buf.String(), "global formatted") {
		t.Errorf("UserInfof() failed to format and output message")
	}
	buf.Reset()

	UserSuccess("global success")
	if !strings.Contains(buf.String(), "global success") {
		t.Errorf("UserSuccess() failed to output message")
	}
	buf.Reset()

	UserSuccessf("global %s", "success")
	if !strings.Contains(buf.String(), "global success") {
		t.Errorf("UserSuccessf() failed to format and output message")
	}
	buf.Reset()

	UserWarning("global warning")
	if !strings.Contains(buf.String(), "global warning") {
		t.Errorf("UserWarning() failed to output message")
	}
	buf.Reset()

	UserWarningf("global %s", "warning")
	if !strings.Contains(buf.String(), "global warning") {
		t.Errorf("UserWarningf() failed to format and output message")
	}
	buf.Reset()

	UserError("global error")
	if !strings.Contains(buf.String(), "global error") {
		t.Errorf("UserError() failed to output message")
	}
	buf.Reset()

	UserErrorf("global %s", "error")
	if !strings.Contains(buf.String(), "global error") {
		t.Errorf("UserErrorf() failed to format and output message")
	}
}

func TestUserOutputDefaultInitialization(t *testing.T) {
	// Reset global output to test initialization
	oldOutput := defaultUserOutput
	defer func() {
		defaultUserOutput = oldOutput
	}()
	defaultUserOutput = nil

	// Should automatically initialize when accessed
	output := GetDefaultUserOutput()
	if output == nil {
		t.Errorf("GetDefaultUserOutput() should initialize automatically")
	}

	// Should be ConsoleOutput type
	if _, ok := output.(*ConsoleOutput); !ok {
		t.Errorf("Default output should be ConsoleOutput type")
	}
}

func TestBackwardCompatibilityThroughLogger(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer
	customOutput := NewConsoleOutputWithWriter(&buf)

	// Set custom output for testing
	oldOutput := defaultUserOutput
	defer func() {
		defaultUserOutput = oldOutput
	}()
	SetDefaultUserOutput(customOutput)

	// Create logger
	_, err := NewLogger(DefaultLogConfig())
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

}
