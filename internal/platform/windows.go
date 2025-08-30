//go:build windows

package platform

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

type windowsProcessManager struct{}

func (w *windowsProcessManager) StartDaemonProcess(execPath string, args []string, logFile *os.File, env []string) (*os.Process, error) {
	// On Windows, we use DETACHED_PROCESS to create a background process
	procAttr := &os.ProcAttr{
		Files: []*os.File{nil, logFile, logFile},
		Env:   env,
		Sys: &syscall.SysProcAttr{
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x00000008, // DETACHED_PROCESS
		},
	}
	
	return os.StartProcess(execPath, args, procAttr)
}

func (w *windowsProcessManager) IsProcessRunning(pid int) bool {
	// Special case for current process
	if pid == os.Getpid() {
		return true
	}
	
	// On Windows, try to find the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	// Try to send a signal to check if process exists
	// On Windows, this approach works for checking process existence
	err = process.Signal(os.Signal(syscall.Signal(0)))
	// If we get "not implemented" error, the process exists but doesn't handle signals
	// If we get other errors, the process likely doesn't exist
	if err != nil {
		// "not implemented" means the process exists but can't handle the signal
		return err.Error() == "not implemented"
	}
	return true
}

func (w *windowsProcessManager) TerminateProcess(pid int) error {
	return w.terminateProcessWindows(pid, false)
}

func (w *windowsProcessManager) KillProcess(pid int) error {
	return w.terminateProcessWindows(pid, true)
}

func (w *windowsProcessManager) terminateProcessWindows(pid int, forceKill bool) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}
	
	// On Windows, we use process.Kill() which terminates the process
	if err := process.Kill(); err != nil {
		return fmt.Errorf("failed to terminate process %d: %w", pid, err)
	}
	
	return nil
}

type windowsSignalHandler struct{}

func (w *windowsSignalHandler) SetupShutdownSignals() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, w.GetShutdownSignals()...)
	return sigChan
}

func (w *windowsSignalHandler) GetShutdownSignals() []os.Signal {
	// Windows supports SIGINT (Ctrl+C) and SIGTERM
	return []os.Signal{os.Interrupt, syscall.SIGTERM}
}

type windowsPathHelper struct{}

func (w *windowsPathHelper) GetDefaultLogPath(filename string) string {
	// Use Windows-appropriate default log path
	appData := os.Getenv("PROGRAMDATA")
	if appData == "" {
		appData = "C:\\ProgramData"
	}
	return filepath.Join(appData, "lai", "logs", filename)
}

func (w *windowsPathHelper) GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".lai"), nil
}

func newPlatform() *Platform {
	return &Platform{
		Process: &windowsProcessManager{},
		Signal:  &windowsSignalHandler{},
		Path:    &windowsPathHelper{},
	}
}