package display

import (
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/shiquda/lai/internal/config"
)

// ColorPrinter provides colored output functionality with graceful degradation
type ColorPrinter struct {
	stdoutColor *color.Color
	stderrColor *color.Color
	enabled     bool
}

// NewColorPrinter creates a new ColorPrinter based on configuration
func NewColorPrinter(cfg config.ColorsConfig) *ColorPrinter {
	cp := &ColorPrinter{
		enabled: cfg.Enabled && isTerminalSupported(),
	}

	if cp.enabled {
		cp.stdoutColor = getColorByName(cfg.Stdout)
		cp.stderrColor = getColorByName(cfg.Stderr)
	}

	return cp
}

// PrintStdout prints text with stdout coloring if enabled
func (cp *ColorPrinter) PrintStdout(text string) string {
	if !cp.enabled || cp.stdoutColor == nil {
		return text
	}
	return cp.stdoutColor.Sprint(text)
}

// PrintStderr prints text with stderr coloring if enabled
func (cp *ColorPrinter) PrintStderr(text string) string {
	if !cp.enabled || cp.stderrColor == nil {
		return text
	}
	return cp.stderrColor.Sprint(text)
}

// IsEnabled returns whether color output is enabled
func (cp *ColorPrinter) IsEnabled() bool {
	return cp.enabled
}

// isTerminalSupported checks if the current terminal supports colors
func isTerminalSupported() bool {
	// Check if output is being redirected
	if !isatty(os.Stdout) || !isatty(os.Stderr) {
		return false
	}

	// Check if explicitly disabled
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check for force color
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Windows-specific handling - don't check TERM variable on Windows
	if runtime.GOOS == "windows" {
		return enableWindowsANSI()
	}

	// On Unix-like systems, check TERM environment variable
	term := strings.ToLower(os.Getenv("TERM"))
	if term == "" || term == "dumb" {
		return false
	}

	// Most modern terminals support color
	return true
}

// isatty checks if the file descriptor is a terminal
func isatty(f *os.File) bool {
	if f == nil {
		return false
	}

	stat, err := f.Stat()
	if err != nil {
		return false
	}

	// Check if it's a character device (typical for terminals)
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// getColorByName returns a color.Color based on color name
func getColorByName(colorName string) *color.Color {
	// Handle empty string by defaulting to gray
	if colorName == "" {
		return color.New(color.FgHiBlack) // Default to gray for empty string
	}

	switch strings.ToLower(colorName) {
	case "black":
		return color.New(color.FgBlack)
	case "red":
		return color.New(color.FgRed)
	case "green":
		return color.New(color.FgGreen)
	case "yellow":
		return color.New(color.FgYellow)
	case "blue":
		return color.New(color.FgBlue)
	case "magenta":
		return color.New(color.FgMagenta)
	case "cyan":
		return color.New(color.FgCyan)
	case "white":
		return color.New(color.FgWhite)
	case "gray", "grey":
		return color.New(color.FgHiBlack) // High intensity black is gray
	case "bright_red":
		return color.New(color.FgHiRed)
	case "bright_green":
		return color.New(color.FgHiGreen)
	case "bright_yellow":
		return color.New(color.FgHiYellow)
	case "bright_blue":
		return color.New(color.FgHiBlue)
	case "bright_magenta":
		return color.New(color.FgHiMagenta)
	case "bright_cyan":
		return color.New(color.FgHiCyan)
	case "bright_white":
		return color.New(color.FgHiWhite)
	default:
		// Default to gray if name is not recognized
		return color.New(color.FgHiBlack)
	}
}

// GetSupportedColors returns a list of supported color names
func GetSupportedColors() []string {
	return []string{
		"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white",
		"gray", "grey", "bright_red", "bright_green", "bright_yellow",
		"bright_blue", "bright_magenta", "bright_cyan", "bright_white",
	}
}
