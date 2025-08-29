package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shiquda/lai/internal/daemon"
)

func TestStopCmd_ProcessNotFound(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_stop_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForStop(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Test that LoadProcessInfo fails for non-existent process
	_, err = testManager.LoadProcessInfo("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent process")
	}
}

func TestStopCmd_ProcessExists(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_stop_exists_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForStop(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create test process with non-existent PID (already stopped)
	testProcess := &daemon.ProcessInfo{
		ID:        "test_stop",
		PID:       99999999, // Non-existent PID
		LogFile:   "/test/app.log",
		StartTime: time.Now(),
		Status:    "running", // Set as running initially
	}

	if err := testManager.SaveProcessInfo(testProcess); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Verify the process exists
	loadedProcess, err := testManager.LoadProcessInfo("test_stop")
	if err != nil {
		t.Fatalf("Process should exist: %v", err)
	}

	if loadedProcess.ID != "test_stop" {
		t.Errorf("Expected process ID 'test_stop', got '%s'", loadedProcess.ID)
	}

	if loadedProcess.PID != 99999999 {
		t.Errorf("Expected PID 99999999, got %d", loadedProcess.PID)
	}
}

func TestStopCmd_StopProcess(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_stop_process_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForStop(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create test process with non-existent PID (will be detected as stopped)
	testProcess := &daemon.ProcessInfo{
		ID:        "test_stop_process",
		PID:       99999998, // Non-existent PID
		LogFile:   "/test/stop.log",
		StartTime: time.Now(),
		Status:    "running",
	}

	if err := testManager.SaveProcessInfo(testProcess); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Test StopProcess - should handle already stopped process gracefully
	err = testManager.StopProcess("test_stop_process")
	if err != nil {
		t.Fatalf("StopProcess should handle stopped process gracefully: %v", err)
	}

	// Verify status is updated to stopped
	updatedProcess, err := testManager.LoadProcessInfo("test_stop_process")
	if err != nil {
		t.Fatalf("Failed to load updated process: %v", err)
	}

	// Note: The status should be updated by the daemon manager's logic
	// Since the PID doesn't exist, it should be marked as stopped
	if updatedProcess.Status == "running" && !testManager.IsProcessRunning(updatedProcess.PID) {
		t.Error("Process status should reflect that it's not actually running")
	}
}

func TestStopCmd_StopAllProcesses(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_stop_all_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForStop(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create multiple test processes
	processes := []*daemon.ProcessInfo{
		{
			ID:        "stop_all_1",
			PID:       99999997,
			LogFile:   "/test/app1.log",
			StartTime: time.Now(),
			Status:    "running",
		},
		{
			ID:        "stop_all_2",
			PID:       99999996,
			LogFile:   "/test/app2.log",
			StartTime: time.Now(),
			Status:    "running",
		},
		{
			ID:        "stop_all_3",
			PID:       99999995,
			LogFile:   "/test/app3.log",
			StartTime: time.Now(),
			Status:    "stopped", // Already stopped
		},
	}

	// Save all processes
	for _, proc := range processes {
		if err := testManager.SaveProcessInfo(proc); err != nil {
			t.Fatalf("Failed to save process %s: %v", proc.ID, err)
		}
	}

	// Test StopAllProcesses
	err = testManager.StopAllProcesses()
	if err != nil {
		t.Fatalf("StopAllProcesses failed: %v", err)
	}

	// Verify all processes exist (they shouldn't be removed, just stopped)
	for _, originalProc := range processes {
		_, err := testManager.LoadProcessInfo(originalProc.ID)
		if err != nil {
			t.Errorf("Process %s should still exist after stopping: %v", originalProc.ID, err)
		}
	}
}

func TestStopCmd_IsProcessRunning(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_stop_running_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForStop(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Test with current process (should be running)
	currentPID := os.Getpid()
	if !testManager.IsProcessRunning(currentPID) {
		t.Error("Current process should be detected as running")
	}

	// Test with non-existent PID (should not be running)
	if testManager.IsProcessRunning(99999994) {
		t.Error("Non-existent process should not be detected as running")
	}
}

func TestStopCmd_ProcessInfoFields(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_stop_fields_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForStop(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create test process with specific values
	startTime := time.Date(2023, 6, 15, 10, 30, 0, 0, time.UTC)
	testProcess := &daemon.ProcessInfo{
		ID:        "field_test_process",
		PID:       12345678,
		LogFile:   "/specific/path/test.log",
		StartTime: startTime,
		Status:    "running",
	}

	if err := testManager.SaveProcessInfo(testProcess); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Load and verify all fields
	loadedProcess, err := testManager.LoadProcessInfo("field_test_process")
	if err != nil {
		t.Fatalf("Failed to load process: %v", err)
	}

	// Verify each field
	if loadedProcess.ID != testProcess.ID {
		t.Errorf("ID mismatch: expected %s, got %s", testProcess.ID, loadedProcess.ID)
	}
	if loadedProcess.PID != testProcess.PID {
		t.Errorf("PID mismatch: expected %d, got %d", testProcess.PID, loadedProcess.PID)
	}
	if loadedProcess.LogFile != testProcess.LogFile {
		t.Errorf("LogFile mismatch: expected %s, got %s", testProcess.LogFile, loadedProcess.LogFile)
	}
	if !loadedProcess.StartTime.Equal(testProcess.StartTime) {
		t.Errorf("StartTime mismatch: expected %v, got %v", testProcess.StartTime, loadedProcess.StartTime)
	}
	if loadedProcess.Status != testProcess.Status {
		t.Errorf("Status mismatch: expected %s, got %s", testProcess.Status, loadedProcess.Status)
	}
}

func TestStopCmd_EmptyProcessList(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_stop_empty_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory (no processes)
	testManager, err := createTestManagerForStop(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Test StopAllProcesses with no processes
	err = testManager.StopAllProcesses()
	if err != nil {
		t.Errorf("StopAllProcesses should handle empty list gracefully: %v", err)
	}

	// Test ListProcesses returns empty list
	processes, err := testManager.ListProcesses()
	if err != nil {
		t.Fatalf("ListProcesses failed: %v", err)
	}

	if len(processes) != 0 {
		t.Errorf("Expected empty process list, got %d processes", len(processes))
	}
}

// Helper function to create a test daemon manager for stop tests
func createTestManagerForStop(tempDir string) (*daemon.Manager, error) {
	processDir := filepath.Join(tempDir, "processes")
	logDir := filepath.Join(tempDir, "logs")

	return daemon.NewManagerWithDirs(processDir, logDir)
}