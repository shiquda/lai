package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shiquda/lai/internal/daemon"
)

func TestResumeProcess_ProcessNotFound(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_resume_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForResume(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Test resuming non-existent process
	err = resumeProcess(testManager, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent process")
	}
	if err != nil && !contains(err.Error(), "process not found") {
		t.Errorf("Expected 'process not found' error, got: %v", err)
	}
}

func TestResumeProcess_AlreadyRunning(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_resume_running_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForResume(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create test process with current PID (running)
	testProcess := &daemon.ProcessInfo{
		ID:        "test_running",
		PID:       os.Getpid(),
		LogFile:   "/test/app.log",
		StartTime: time.Now(),
		Status:    "running",
	}

	if err := testManager.SaveProcessInfo(testProcess); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Test resuming already running process
	err = resumeProcess(testManager, "test_running")
	if err == nil {
		t.Error("Expected error for already running process")
	}
	if err != nil && !contains(err.Error(), "already running") {
		t.Errorf("Expected 'already running' error, got: %v", err)
	}
}

func TestContainsTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		processID string
		expected  bool
	}{
		{
			name:      "Process ID with timestamp",
			processID: "webapp_1234567890",
			expected:  true,
		},
		{
			name:      "Process ID with multiple underscores and digits",
			processID: "api_service_9876543210",
			expected:  true,
		},
		{
			name:      "Process ID without timestamp",
			processID: "webapp",
			expected:  false,
		},
		{
			name:      "Process ID with underscore but no digits",
			processID: "webapp_service",
			expected:  false,
		},
		{
			name:      "Process ID with digits but no underscore",
			processID: "webapp123",
			expected:  false,
		},
		{
			name:      "Empty process ID",
			processID: "",
			expected:  false,
		},
		{
			name:      "Process ID with underscore at end",
			processID: "webapp_",
			expected:  false,
		},
		{
			name:      "Process ID with underscore followed by mixed chars",
			processID: "webapp_123abc",
			expected:  true, // Has digits after underscore
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsTimestamp(tt.processID)
			if result != tt.expected {
				t.Errorf("containsTimestamp(%s) = %v, expected %v",
					tt.processID, result, tt.expected)
			}
		})
	}
}

func TestResumeProcess_StoppedProcess(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_resume_stopped_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForResume(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create test process with non-existent PID (stopped)
	testProcess := &daemon.ProcessInfo{
		ID:        "test_stopped",
		PID:       99999999, // Non-existent PID
		LogFile:   "/test/app.log",
		StartTime: time.Now(),
		Status:    "stopped",
	}

	if err := testManager.SaveProcessInfo(testProcess); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Note: We can't fully test the actual process spawning in unit tests
	// as it would require forking and running the actual lai binary.
	// Instead, we test the logic up to the point of process creation.

	// Test that the process exists and is stopped
	info, err := testManager.LoadProcessInfo("test_stopped")
	if err != nil {
		t.Errorf("Process should exist: %v", err)
	}

	if testManager.IsProcessRunning(info.PID) {
		t.Error("Process should not be running")
	}

	// We can't test the actual resumeProcess function completely without
	// mocking os.StartProcess, but we can test that it attempts to
	// load the process info correctly
}

func TestResumeProcess_ProcessInfoHandling(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_resume_info_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForResume(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create test process
	testProcess := &daemon.ProcessInfo{
		ID:        "test_info_handling",
		PID:       99999998,
		LogFile:   "/test/specific.log",
		StartTime: time.Now(),
		Status:    "stopped",
	}

	if err := testManager.SaveProcessInfo(testProcess); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Verify the process info can be loaded correctly
	loadedInfo, err := testManager.LoadProcessInfo("test_info_handling")
	if err != nil {
		t.Fatalf("Failed to load process info: %v", err)
	}

	// Verify all fields match
	if loadedInfo.ID != testProcess.ID {
		t.Errorf("ID mismatch: expected %s, got %s", testProcess.ID, loadedInfo.ID)
	}
	if loadedInfo.LogFile != testProcess.LogFile {
		t.Errorf("LogFile mismatch: expected %s, got %s", testProcess.LogFile, loadedInfo.LogFile)
	}
	if loadedInfo.PID != testProcess.PID {
		t.Errorf("PID mismatch: expected %d, got %d", testProcess.PID, loadedInfo.PID)
	}

	// Test log path generation
	expectedLogPath := filepath.Join(tempDir, "logs", "test_info_handling.log")
	actualLogPath := testManager.GetProcessLogPath("test_info_handling")
	if actualLogPath != expectedLogPath {
		t.Errorf("Log path mismatch: expected %s, got %s", expectedLogPath, actualLogPath)
	}
}

func TestResumeProcess_CustomNameDetection(t *testing.T) {
	// Test the logic for detecting custom names vs timestamp-based IDs
	tests := []struct {
		processID    string
		expectCustom bool
		description  string
	}{
		{
			processID:    "webapp",
			expectCustom: true,
			description:  "Simple custom name without timestamp",
		},
		{
			processID:    "my_service",
			expectCustom: true,
			description:  "Custom name with underscore but no timestamp",
		},
		{
			processID:    "app_123456",
			expectCustom: false,
			description:  "Generated ID with timestamp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			hasTimestamp := containsTimestamp(tt.processID)
			isCustom := !hasTimestamp

			if isCustom != tt.expectCustom {
				t.Errorf("Expected custom name detection %v for ID %s, got %v",
					tt.expectCustom, tt.processID, isCustom)
			}
		})
	}
}

// Helper function to create a test daemon manager for resume tests
func createTestManagerForResume(tempDir string) (*daemon.Manager, error) {
	processDir := filepath.Join(tempDir, "processes")
	logDir := filepath.Join(tempDir, "logs")

	return daemon.NewManagerWithDirs(processDir, logDir)
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				strings.Contains(s, substr))))
}
