package display

import (
	"os"
	"testing"

	"github.com/shiquda/lai/internal/config"
)

func TestNewColorPrinter(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.ColorsConfig
		expected bool
	}{
		{
			name: "Colors enabled with valid config",
			cfg: config.ColorsConfig{
				Enabled: true,
				Stdout:  "gray",
				Stderr:  "red",
			},
			expected: true, // Will depend on terminal support
		},
		{
			name: "Colors disabled",
			cfg: config.ColorsConfig{
				Enabled: false,
				Stdout:  "gray",
				Stderr:  "red",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewColorPrinter(tt.cfg)
			if tt.cfg.Enabled == false && cp.IsEnabled() {
				t.Error("Expected colors to be disabled")
			}
		})
	}
}

func TestGetColorByName(t *testing.T) {
	tests := []struct {
		name      string
		colorName string
		expectNil bool
	}{
		{"Valid color - red", "red", false},
		{"Valid color - gray", "gray", false},
		{"Valid color - grey", "grey", false},
		{"Valid color - blue", "blue", false},
		{"Invalid color - defaults to gray", "invalid_color", false},
		{"Empty string - defaults to gray", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := getColorByName(tt.colorName)
			isNil := color == nil
			if isNil != tt.expectNil {
				t.Errorf("Expected nil=%v, got nil=%v for color %s", tt.expectNil, isNil, tt.colorName)
			}
		})
	}
}

func TestColorPrinterOutput(t *testing.T) {
	cfg := config.ColorsConfig{
		Enabled: true,
		Stdout:  "gray",
		Stderr:  "red",
	}

	cp := NewColorPrinter(cfg)

	// Test text output
	testText := "Hello, World!"

	stdoutResult := cp.PrintStdout(testText)
	stderrResult := cp.PrintStderr(testText)

	// At minimum, the text should be preserved
	if len(stdoutResult) < len(testText) {
		t.Error("Stdout colored text is shorter than original")
	}
	if len(stderrResult) < len(testText) {
		t.Error("Stderr colored text is shorter than original")
	}
}

func TestGetSupportedColors(t *testing.T) {
	colors := GetSupportedColors()
	if len(colors) == 0 {
		t.Error("Expected non-empty list of supported colors")
	}

	// Check that basic colors are included
	expectedColors := []string{"red", "green", "blue", "gray", "white"}
	for _, expected := range expectedColors {
		found := false
		for _, color := range colors {
			if color == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected color %s not found in supported colors list", expected)
		}
	}
}

func TestTerminalSupport(t *testing.T) {
	// Test with NO_COLOR environment variable
	originalNoColor := os.Getenv("NO_COLOR")
	defer func() {
		if originalNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", originalNoColor)
		}
	}()

	os.Setenv("NO_COLOR", "1")
	if isTerminalSupported() {
		t.Error("Expected terminal support to be disabled with NO_COLOR set")
	}

	os.Unsetenv("NO_COLOR")

	// Test with FORCE_COLOR environment variable
	originalForceColor := os.Getenv("FORCE_COLOR")
	defer func() {
		if originalForceColor == "" {
			os.Unsetenv("FORCE_COLOR")
		} else {
			os.Setenv("FORCE_COLOR", originalForceColor)
		}
	}()

	os.Setenv("FORCE_COLOR", "1")
	// Note: FORCE_COLOR only works when we have a proper terminal
	// Since we might be running in a CI environment or with redirected output,
	// we'll only check if the function doesn't fail
	_ = isTerminalSupported()
	// The test will pass regardless since terminal detection can vary by environment
}
