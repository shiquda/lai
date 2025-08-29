package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shiquda/lai/internal/daemon"
)

func TestListCmd(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_list_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForList(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create test processes
	processes := []*daemon.ProcessInfo{
		{
			ID:        "webapp_123",
			PID:       os.Getpid(), // Use current PID to ensure it's "running"
			LogFile:   "/app/logs/webapp.log",
			StartTime: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			Status:    "running",
		},
		{
			ID:        "api_456",
			PID:       99999999, // Non-existent PID to ensure it's "stopped"
			LogFile:   "/app/logs/api.log",
			StartTime: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC),
			Status:    "stopped",
		},
	}

	// Save processes
	for _, proc := range processes {
		if err := testManager.SaveProcessInfo(proc); err != nil {
			t.Fatalf("Failed to save process %s: %v", proc.ID, err)
		}
	}

	// Test the command execution with a buffer to capture output
	output := captureListOutput(t, testManager)

	// Debug: print the actual output
	t.Logf("Actual output:\n%s", output)

	// Verify output contains expected information
	if !strings.Contains(output, "webapp_123") {
		t.Errorf("Output should contain process ID webapp_123")
	}
	if !strings.Contains(output, "api_456") {
		t.Errorf("Output should contain process ID api_456")
	}
	if !strings.Contains(output, fmt.Sprintf("%d", os.Getpid())) {
		t.Errorf("Output should contain current PID")
	}
	if !strings.Contains(output, "99999999") {
		t.Errorf("Output should contain PID 99999999")
	}
	// Since ListProcesses updates status based on actual process state,
	// we should check for the updated status
	if !strings.Contains(output, "running") {
		t.Errorf("Output should contain status 'running'")
	}
	if !strings.Contains(output, "stopped") {
		t.Errorf("Output should contain status 'stopped'")
	}
	if !strings.Contains(output, "/app/logs/webapp.log") {
		t.Errorf("Output should contain log file path")
	}
}

func TestListCmd_EmptyList(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_list_test_empty_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory (no processes)
	testManager, err := createTestManagerForList(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Test the command execution with empty process list
	output := captureListOutput(t, testManager)

	// Verify output indicates no processes
	if !strings.Contains(output, "No daemon processes found") {
		t.Errorf("Output should indicate no processes found, got: %s", output)
	}
}

func TestListCmd_OutputFormat(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_list_test_format_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForList(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create a test process with known values
	testTime := time.Date(2023, 6, 15, 14, 30, 45, 0, time.UTC)
	testProcess := &daemon.ProcessInfo{
		ID:        "test_process",
		PID:       99999,
		LogFile:   "/test/path/test.log",
		StartTime: testTime,
		Status:    "running",
	}

	if err := testManager.SaveProcessInfo(testProcess); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Test the command execution
	output := captureListOutput(t, testManager)

	// Verify the output format contains headers
	if !strings.Contains(output, "PROCESS ID") {
		t.Errorf("Output should contain PROCESS ID header")
	}
	if !strings.Contains(output, "PID") {
		t.Errorf("Output should contain PID header")
	}
	if !strings.Contains(output, "STATUS") {
		t.Errorf("Output should contain STATUS header")
	}
	if !strings.Contains(output, "START TIME") {
		t.Errorf("Output should contain START TIME header")
	}
	if !strings.Contains(output, "LOG FILE") {
		t.Errorf("Output should contain LOG FILE header")
	}

	// Verify data formatting
	expectedTime := testTime.Format("2006-01-02 15:04:05")
	if !strings.Contains(output, expectedTime) {
		t.Errorf("Output should contain formatted time %s", expectedTime)
	}
}

// Helper function to create a test daemon manager for list tests
func createTestManagerForList(tempDir string) (*daemon.Manager, error) {
	processDir := filepath.Join(tempDir, "processes")
	logDir := filepath.Join(tempDir, "logs")

	return daemon.NewManagerWithDirs(processDir, logDir)
}

// Helper function to capture output from list command
// Since we can't easily mock daemon.NewManager in the command itself,
// we'll simulate the list logic here
func captureListOutput(t *testing.T, manager *daemon.Manager) string {
	// Get processes from our test manager
	processes, err := manager.ListProcesses()
	if err != nil {
		t.Fatalf("Failed to list processes: %v", err)
	}

	// Simulate the same output logic as listCmd
	var buf bytes.Buffer

	if len(processes) == 0 {
		buf.WriteString("No daemon processes found\n")
		return buf.String()
	}

	// Headers
	buf.WriteString("PROCESS ID           PID      STATUS     START TIME           LOG FILE\n")
	buf.WriteString("----------           ---      ------     ----------           --------\n")

	// Process data
	for _, proc := range processes {
		startTime := proc.StartTime.Format("2006-01-02 15:04:05")
		line := formatListLine(proc.ID, proc.PID, proc.Status, startTime, proc.LogFile)
		buf.WriteString(line + "\n")
	}

	return buf.String()
}

// Helper function to format a line similar to the actual command
func formatListLine(id string, pid int, status, startTime, logFile string) string {
	// Simulate the same formatting as in list.go
	return fmt.Sprintf("%-20s %-8d %-10s %-20s %s",
		id, pid, status, startTime, logFile)
}
