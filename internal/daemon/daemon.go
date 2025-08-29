package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

// ProcessInfo represents information about a running daemon process
type ProcessInfo struct {
	ID        string    `json:"id"`
	PID       int       `json:"pid"`
	LogFile   string    `json:"log_file"`
	StartTime time.Time `json:"start_time"`
	Status    string    `json:"status"`
}

// Manager handles daemon process management
type Manager struct {
	processDir string
	logDir     string
}

// NewManager creates a new daemon manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	processDir := filepath.Join(homeDir, ".lai", "processes")
	logDir := filepath.Join(homeDir, ".lai", "logs")

	// Ensure directories exist
	if err := os.MkdirAll(processDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create process directory: %w", err)
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &Manager{
		processDir: processDir,
		logDir:     logDir,
	}, nil
}

// NewManagerWithDirs creates a new daemon manager with custom directories (for testing)
func NewManagerWithDirs(processDir, logDir string) (*Manager, error) {
	// Ensure directories exist
	if err := os.MkdirAll(processDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create process directory: %w", err)
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	return &Manager{
		processDir: processDir,
		logDir:     logDir,
	}, nil
}

// GenerateProcessID generates a unique process ID
func (m *Manager) GenerateProcessID(logFile string) string {
	timestamp := time.Now().Unix()
	basename := filepath.Base(logFile)
	return fmt.Sprintf("%s_%d", basename, timestamp)
}

// GenerateProcessIDWithName generates a process ID with custom name
func (m *Manager) GenerateProcessIDWithName(name string) string {
	return name
}

// ProcessExists checks if a process with the given ID already exists
func (m *Manager) ProcessExists(processID string) bool {
	_, err := m.LoadProcessInfo(processID)
	return err == nil
}

// SaveProcessInfo saves process information to file
func (m *Manager) SaveProcessInfo(info *ProcessInfo) error {
	filePath := filepath.Join(m.processDir, info.ID+".json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal process info: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write process info: %w", err)
	}

	return nil
}

// LoadProcessInfo loads process information from file
func (m *Manager) LoadProcessInfo(processID string) (*ProcessInfo, error) {
	filePath := filepath.Join(m.processDir, processID+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read process info: %w", err)
	}

	var info ProcessInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to unmarshal process info: %w", err)
	}

	return &info, nil
}

// ListProcesses lists all daemon processes
func (m *Manager) ListProcesses() ([]*ProcessInfo, error) {
	files, err := filepath.Glob(filepath.Join(m.processDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list process files: %w", err)
	}

	var processes []*ProcessInfo
	for _, file := range files {
		basename := filepath.Base(file)
		processID := basename[:len(basename)-5] // Remove .json extension

		info, err := m.LoadProcessInfo(processID)
		if err != nil {
			continue // Skip invalid files
		}

		// Check if process is still running
		info.Status = m.getProcessStatus(info.PID)
		
		// Update the file with current status
		m.SaveProcessInfo(info)

		processes = append(processes, info)
	}

	return processes, nil
}

// RemoveProcessInfo removes process information file
func (m *Manager) RemoveProcessInfo(processID string) error {
	filePath := filepath.Join(m.processDir, processID+".json")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove process info: %w", err)
	}
	return nil
}

// GetProcessLogPath returns the log file path for a process
func (m *Manager) GetProcessLogPath(processID string) string {
	return filepath.Join(m.logDir, processID+".log")
}

// StopProcess stops a daemon process
func (m *Manager) StopProcess(processID string) error {
	info, err := m.LoadProcessInfo(processID)
	if err != nil {
		return fmt.Errorf("failed to load process info: %w", err)
	}

	// Check if process is still running
	if !m.isProcessRunning(info.PID) {
		// Process already stopped, just update status
		info.Status = "stopped"
		return m.SaveProcessInfo(info)
	}

	// Send SIGTERM to the process
	process, err := os.FindProcess(info.PID)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", info.PID, err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to process %d: %w", info.PID, err)
	}

	// Wait a bit and check if process stopped
	time.Sleep(2 * time.Second)
	if !m.isProcessRunning(info.PID) {
		info.Status = "stopped"
		return m.SaveProcessInfo(info)
	}

	// If still running, force kill
	if err := process.Signal(syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to send SIGKILL to process %d: %w", info.PID, err)
	}

	// Update status to stopped
	info.Status = "stopped"
	return m.SaveProcessInfo(info)
}

// StopAllProcesses stops all daemon processes
func (m *Manager) StopAllProcesses() error {
	processes, err := m.ListProcesses()
	if err != nil {
		return err
	}

	var errors []error
	for _, proc := range processes {
		if err := m.StopProcess(proc.ID); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop process %s: %w", proc.ID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to stop some processes: %v", errors)
	}

	return nil
}

// IsProcessRunning checks if a process is still running (public method)
func (m *Manager) IsProcessRunning(pid int) bool {
	return m.isProcessRunning(pid)
}

// isProcessRunning checks if a process is still running
func (m *Manager) isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// getProcessStatus returns the status of a process
func (m *Manager) getProcessStatus(pid int) string {
	if m.isProcessRunning(pid) {
		return "running"
	}
	return "stopped"
}

// Daemonize creates a daemon process
func Daemonize() error {
	// Fork the process
	pid, err := syscall.ForkExec(os.Args[0], os.Args, &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{0, 1, 2}, // stdin, stdout, stderr
		Sys:   &syscall.SysProcAttr{Setsid: true},
	})

	if err != nil {
		return fmt.Errorf("failed to fork process: %w", err)
	}

	// Parent process exits
	fmt.Printf("Daemon started with PID: %d\n", pid)
	os.Exit(0)

	// This code won't be reached by the parent process
	return nil
}

// CreatePidFile creates a PID file for the daemon
func CreatePidFile(pidFile string) error {
	pid := os.Getpid()
	return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}

// RemovePidFile removes the PID file
func RemovePidFile(pidFile string) error {
	return os.Remove(pidFile)
}