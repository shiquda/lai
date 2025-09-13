//go:build windows

package display

import (
	"os"

	"golang.org/x/sys/windows"
)

// enableWindowsANSI enables ANSI color support on Windows 10+
func enableWindowsANSI() bool {
	// Try to enable ANSI support on Windows console
	var mode uint32

	// Handle stdout
	stdout := windows.Handle(os.Stdout.Fd())
	if err := windows.GetConsoleMode(stdout, &mode); err != nil {
		// Not a console, probably redirected output, assume color support
		return true
	}

	// Check if virtual terminal processing is already enabled
	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING == 0 {
		// Try to enable virtual terminal processing
		mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
		if err := windows.SetConsoleMode(stdout, mode); err != nil {
			// Failed to enable, but we'll still allow color output
			// Some Windows terminals support ANSI without explicit enabling
			return true
		}
	}

	// Handle stderr
	stderr := windows.Handle(os.Stderr.Fd())
	if err := windows.GetConsoleMode(stderr, &mode); err == nil {
		if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING == 0 {
			mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
			windows.SetConsoleMode(stderr, mode)
		}
	}

	return true
}
