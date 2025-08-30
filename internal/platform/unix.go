//go:build !windows

package platform

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type unixProcessManager struct{}

func (u *unixProcessManager) StartDaemonProcess(execPath string, args []string, logFile *os.File, env []string) (*os.Process, error) {
	procAttr := &os.ProcAttr{
		Files: []*os.File{nil, logFile, logFile},
		Env:   env,
		Sys:   &syscall.SysProcAttr{Setsid: true},
	}
	
	return os.StartProcess(execPath, args, procAttr)
}

func (u *unixProcessManager) IsProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (u *unixProcessManager) TerminateProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}
	
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to process %d: %w", pid, err)
	}
	
	// Wait a bit to see if process stops gracefully
	time.Sleep(2 * time.Second)
	if !u.IsProcessRunning(pid) {
		return nil
	}
	
	// Process still running, return without force killing
	// Let the caller decide whether to force kill
	return fmt.Errorf("process %d did not terminate gracefully", pid)
}

func (u *unixProcessManager) KillProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}
	
	if err := process.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to send SIGKILL to process %d: %w", pid, err)
	}
	
	return nil
}

type unixSignalHandler struct{}

func (u *unixSignalHandler) SetupShutdownSignals() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, u.GetShutdownSignals()...)
	return sigChan
}

func (u *unixSignalHandler) GetShutdownSignals() []os.Signal {
	return []os.Signal{syscall.SIGINT, syscall.SIGTERM}
}

type unixPathHelper struct{}

func (u *unixPathHelper) GetDefaultLogPath(filename string) string {
	return filepath.Join("/var/log", filename)
}

func (u *unixPathHelper) GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".lai"), nil
}

func newPlatform() *Platform {
	return &Platform{
		Process: &unixProcessManager{},
		Signal:  &unixSignalHandler{},
		Path:    &unixPathHelper{},
	}
}