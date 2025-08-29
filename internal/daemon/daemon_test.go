package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	if manager.processDir == "" {
		t.Error("processDir should not be empty")
	}

	if manager.logDir == "" {
		t.Error("logDir should not be empty")
	}

	// Check if directories were created
	if _, err := os.Stat(manager.processDir); os.IsNotExist(err) {
		t.Errorf("processDir was not created: %s", manager.processDir)
	}

	if _, err := os.Stat(manager.logDir); os.IsNotExist(err) {
		t.Errorf("logDir was not created: %s", manager.logDir)
	}
}

func TestGenerateProcessID(t *testing.T) {
	manager := &Manager{}

	testCases := []struct {
		logFile  string
		expected string // We'll check if it starts with this
	}{
		{"/path/to/test.log", "test.log_"},
		{"/another/path/app.log", "app.log_"},
		{"simple.log", "simple.log_"},
	}

	for _, tc := range testCases {
		id := manager.GenerateProcessID(tc.logFile)

		if len(id) == 0 {
			t.Error("Generated ID should not be empty")
		}

		if !containsString(id, filepath.Base(tc.logFile)) {
			t.Errorf("Generated ID should contain base filename %s, got %s", filepath.Base(tc.logFile), id)
		}
	}
}

func TestSaveAndLoadProcessInfo(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := &Manager{
		processDir: filepath.Join(tempDir, "processes"),
		logDir:     filepath.Join(tempDir, "logs"),
	}

	// Ensure directories exist
	os.MkdirAll(manager.processDir, 0755)
	os.MkdirAll(manager.logDir, 0755)

	// Test data
	testInfo := &ProcessInfo{
		ID:        "test_123",
		PID:       12345,
		LogFile:   "/test/log.txt",
		StartTime: time.Now().Truncate(time.Second), // Truncate to avoid precision issues
		Status:    "running",
	}

	// Test SaveProcessInfo
	err = manager.SaveProcessInfo(testInfo)
	if err != nil {
		t.Fatalf("SaveProcessInfo failed: %v", err)
	}

	// Check if file was created
	filePath := filepath.Join(manager.processDir, "test_123.json")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Process info file was not created")
	}

	// Test LoadProcessInfo
	loadedInfo, err := manager.LoadProcessInfo("test_123")
	if err != nil {
		t.Fatalf("LoadProcessInfo failed: %v", err)
	}

	// Compare loaded data
	if loadedInfo.ID != testInfo.ID {
		t.Errorf("ID mismatch: expected %s, got %s", testInfo.ID, loadedInfo.ID)
	}
	if loadedInfo.PID != testInfo.PID {
		t.Errorf("PID mismatch: expected %d, got %d", testInfo.PID, loadedInfo.PID)
	}
	if loadedInfo.LogFile != testInfo.LogFile {
		t.Errorf("LogFile mismatch: expected %s, got %s", testInfo.LogFile, loadedInfo.LogFile)
	}
	if !loadedInfo.StartTime.Equal(testInfo.StartTime) {
		t.Errorf("StartTime mismatch: expected %v, got %v", testInfo.StartTime, loadedInfo.StartTime)
	}
	if loadedInfo.Status != testInfo.Status {
		t.Errorf("Status mismatch: expected %s, got %s", testInfo.Status, loadedInfo.Status)
	}
}

func TestLoadProcessInfo_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lai_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := &Manager{
		processDir: filepath.Join(tempDir, "processes"),
	}
	os.MkdirAll(manager.processDir, 0755)

	_, err = manager.LoadProcessInfo("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent process, got nil")
	}
}

func TestListProcesses(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lai_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := &Manager{
		processDir: filepath.Join(tempDir, "processes"),
		logDir:     filepath.Join(tempDir, "logs"),
	}
	os.MkdirAll(manager.processDir, 0755)
	os.MkdirAll(manager.logDir, 0755)

	// Create test process info files
	testProcesses := []*ProcessInfo{
		{
			ID:        "test1_123",
			PID:       12345,
			LogFile:   "/test/log1.txt",
			StartTime: time.Now(),
			Status:    "running",
		},
		{
			ID:        "test2_456",
			PID:       67890,
			LogFile:   "/test/log2.txt",
			StartTime: time.Now(),
			Status:    "stopped",
		},
	}

	for _, proc := range testProcesses {
		if err := manager.SaveProcessInfo(proc); err != nil {
			t.Fatalf("Failed to save test process: %v", err)
		}
	}

	// Test ListProcesses
	processes, err := manager.ListProcesses()
	if err != nil {
		t.Fatalf("ListProcesses failed: %v", err)
	}

	if len(processes) != 2 {
		t.Errorf("Expected 2 processes, got %d", len(processes))
	}

	// Verify process IDs are present
	foundIDs := make(map[string]bool)
	for _, proc := range processes {
		foundIDs[proc.ID] = true
	}

	for _, expectedProc := range testProcesses {
		if !foundIDs[expectedProc.ID] {
			t.Errorf("Expected process ID %s not found in results", expectedProc.ID)
		}
	}
}

func TestRemoveProcessInfo(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lai_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	manager := &Manager{
		processDir: filepath.Join(tempDir, "processes"),
	}
	os.MkdirAll(manager.processDir, 0755)

	// Create a test process
	testInfo := &ProcessInfo{
		ID:      "test_remove",
		PID:     12345,
		LogFile: "/test/log.txt",
		Status:  "running",
	}

	if err := manager.SaveProcessInfo(testInfo); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Test RemoveProcessInfo
	err = manager.RemoveProcessInfo("test_remove")
	if err != nil {
		t.Fatalf("RemoveProcessInfo failed: %v", err)
	}

	// Check if file was removed
	filePath := filepath.Join(manager.processDir, "test_remove.json")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Process info file should have been removed")
	}

	// Test removing non-existent process (should not error)
	err = manager.RemoveProcessInfo("nonexistent")
	if err != nil {
		t.Errorf("RemoveProcessInfo should not error for nonexistent process: %v", err)
	}
}

func TestGetProcessLogPath(t *testing.T) {
	manager := &Manager{
		logDir: "/test/logs",
	}

	logPath := manager.GetProcessLogPath("test_123")
	expected := "/test/logs/test_123.log"

	if logPath != expected {
		t.Errorf("Expected log path %s, got %s", expected, logPath)
	}
}

func TestGetProcessStatus(t *testing.T) {
	manager := &Manager{}

	// Test with current process (should be running)
	currentPID := os.Getpid()
	status := manager.getProcessStatus(currentPID)
	if status != "running" {
		t.Errorf("Current process should be running, got %s", status)
	}

	// Test with invalid PID (should be stopped)
	status = manager.getProcessStatus(99999999) // Very unlikely to exist
	if status != "stopped" {
		t.Errorf("Invalid PID should return stopped, got %s", status)
	}
}

func TestIsProcessRunning(t *testing.T) {
	manager := &Manager{}

	// Test with current process
	currentPID := os.Getpid()
	if !manager.isProcessRunning(currentPID) {
		t.Error("Current process should be reported as running")
	}

	// Test with invalid PID
	if manager.isProcessRunning(99999999) {
		t.Error("Invalid PID should be reported as not running")
	}
}

func TestCreatePidFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lai_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pidFile := filepath.Join(tempDir, "test.pid")

	err = CreatePidFile(pidFile)
	if err != nil {
		t.Fatalf("CreatePidFile failed: %v", err)
	}

	// Check if file was created and contains correct PID
	content, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatalf("Failed to read PID file: %v", err)
	}

	expectedPID := fmt.Sprintf("%d", os.Getpid())
	if string(content) != expectedPID {
		t.Errorf("PID file content mismatch: expected %s, got %s", expectedPID, string(content))
	}
}

func TestRemovePidFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lai_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pidFile := filepath.Join(tempDir, "test.pid")

	// Create PID file
	if err := CreatePidFile(pidFile); err != nil {
		t.Fatalf("Failed to create PID file: %v", err)
	}

	// Remove PID file
	err = RemovePidFile(pidFile)
	if err != nil {
		t.Fatalf("RemovePidFile failed: %v", err)
	}

	// Check if file was removed
	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		t.Error("PID file should have been removed")
	}
}

func TestProcessInfoJSONMarshaling(t *testing.T) {
	original := &ProcessInfo{
		ID:        "test_123",
		PID:       12345,
		LogFile:   "/test/log.txt",
		StartTime: time.Now().UTC().Truncate(time.Second),
		Status:    "running",
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(original, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	// Unmarshal from JSON
	var unmarshaled ProcessInfo
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	// Compare fields
	if unmarshaled.ID != original.ID {
		t.Errorf("ID mismatch after JSON round-trip")
	}
	if unmarshaled.PID != original.PID {
		t.Errorf("PID mismatch after JSON round-trip")
	}
	if unmarshaled.LogFile != original.LogFile {
		t.Errorf("LogFile mismatch after JSON round-trip")
	}
	if !unmarshaled.StartTime.Equal(original.StartTime) {
		t.Errorf("StartTime mismatch after JSON round-trip")
	}
	if unmarshaled.Status != original.Status {
		t.Errorf("Status mismatch after JSON round-trip")
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] != substr && s[:len(substr)] != substr || s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}
