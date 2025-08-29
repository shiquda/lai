package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shiquda/lai/internal/daemon"
)

func TestCleanSingleProcess(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_clean_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}
	
	// Create a test process
	testInfo := &daemon.ProcessInfo{
		ID:        "test_clean",
		PID:       99999999, // Non-existent PID (stopped)
		LogFile:   "/test/log.txt",
		StartTime: time.Now(),
		Status:    "stopped",
	}

	// Save test process
	if err := testManager.SaveProcessInfo(testInfo); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Test cleanSingleProcess with stopped process
	err = cleanSingleProcess(testManager, "test_clean")
	if err != nil {
		t.Fatalf("cleanSingleProcess failed: %v", err)
	}

	// Verify process was removed
	_, err = testManager.LoadProcessInfo("test_clean")
	if err == nil {
		t.Error("Process should have been removed")
	}
}

func TestCleanSingleProcess_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lai_clean_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testManager, err := createTestManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Test cleaning non-existent process
	err = cleanSingleProcess(testManager, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent process")
	}
}

func TestCleanSingleProcess_StillRunning(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lai_clean_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testManager, err := createTestManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create a test process with current PID (running)
	testInfo := &daemon.ProcessInfo{
		ID:        "test_running",
		PID:       os.Getpid(),
		LogFile:   "/test/log.txt",
		StartTime: time.Now(),
		Status:    "running",
	}

	if err := testManager.SaveProcessInfo(testInfo); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Test cleaning running process (should fail)
	err = cleanSingleProcess(testManager, "test_running")
	if err == nil {
		t.Error("Expected error when trying to clean running process")
	}
}

func TestCleanAllStoppedProcesses(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lai_clean_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testManager, err := createTestManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create test processes
	processes := []*daemon.ProcessInfo{
		{
			ID:        "stopped1",
			PID:       99999998,
			LogFile:   "/test/log1.txt",
			StartTime: time.Now(),
			Status:    "stopped",
		},
		{
			ID:        "stopped2",
			PID:       99999997,
			LogFile:   "/test/log2.txt",
			StartTime: time.Now(),
			Status:    "stopped",
		},
		{
			ID:        "running1",
			PID:       os.Getpid(),
			LogFile:   "/test/log3.txt",
			StartTime: time.Now(),
			Status:    "running",
		},
	}

	// Save all processes
	for _, proc := range processes {
		if err := testManager.SaveProcessInfo(proc); err != nil {
			t.Fatalf("Failed to save process %s: %v", proc.ID, err)
		}
	}

	// Test cleanAllStoppedProcesses
	err = cleanAllStoppedProcesses(testManager)
	if err != nil {
		t.Fatalf("cleanAllStoppedProcesses failed: %v", err)
	}

	// Verify stopped processes were removed
	_, err = testManager.LoadProcessInfo("stopped1")
	if err == nil {
		t.Error("Stopped process 1 should have been removed")
	}

	_, err = testManager.LoadProcessInfo("stopped2")
	if err == nil {
		t.Error("Stopped process 2 should have been removed")
	}

	// Verify running process was not removed
	_, err = testManager.LoadProcessInfo("running1")
	if err != nil {
		t.Error("Running process should not have been removed")
	}
}

func TestCleanAllStoppedProcesses_NothingToClean(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lai_clean_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testManager, err := createTestManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create only running processes
	testInfo := &daemon.ProcessInfo{
		ID:        "running_only",
		PID:       os.Getpid(),
		LogFile:   "/test/log.txt",
		StartTime: time.Now(),
		Status:    "running",
	}

	if err := testManager.SaveProcessInfo(testInfo); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Test cleaning with no stopped processes
	err = cleanAllStoppedProcesses(testManager)
	if err != nil {
		t.Fatalf("cleanAllStoppedProcesses should not fail when nothing to clean: %v", err)
	}

	// Verify running process still exists
	_, err = testManager.LoadProcessInfo("running_only")
	if err != nil {
		t.Error("Running process should still exist")
	}
}

func TestCleanAllStoppedProcesses_EmptyList(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lai_clean_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testManager, err := createTestManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Test cleaning with empty process list
	err = cleanAllStoppedProcesses(testManager)
	if err != nil {
		t.Fatalf("cleanAllStoppedProcesses should not fail with empty list: %v", err)
	}
}

// Helper function to create a test daemon manager
func createTestManager(tempDir string) (*daemon.Manager, error) {
	processDir := filepath.Join(tempDir, "processes")
	logDir := filepath.Join(tempDir, "logs")

	return daemon.NewManagerWithDirs(processDir, logDir)
}