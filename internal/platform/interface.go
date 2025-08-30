package platform

import (
	"os"
)

// ProcessManager provides cross-platform process management
type ProcessManager interface {
	// StartDaemonProcess starts a process in background mode
	StartDaemonProcess(execPath string, args []string, logFile *os.File, env []string) (*os.Process, error)
	
	// IsProcessRunning checks if a process with given PID is running
	IsProcessRunning(pid int) bool
	
	// TerminateProcess terminates a process gracefully
	TerminateProcess(pid int) error
	
	// KillProcess force kills a process
	KillProcess(pid int) error
}

// SignalHandler provides cross-platform signal handling
type SignalHandler interface {
	// SetupShutdownSignals sets up signal handling for graceful shutdown
	SetupShutdownSignals() chan os.Signal
	
	// GetShutdownSignals returns the signals to listen for shutdown
	GetShutdownSignals() []os.Signal
}

// PathHelper provides cross-platform path utilities
type PathHelper interface {
	// GetDefaultLogPath returns the default log file path for the platform
	GetDefaultLogPath(filename string) string
	
	// GetConfigDir returns the configuration directory for the platform
	GetConfigDir() (string, error)
}

// Platform provides all platform-specific functionality
type Platform struct {
	Process ProcessManager
	Signal  SignalHandler
	Path    PathHelper
}

// New returns a platform instance for the current OS
func New() *Platform {
	return newPlatform()
}