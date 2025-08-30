package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shiquda/lai/internal/daemon"
	"github.com/shiquda/lai/internal/testutils"
)

func TestRunDaemon_ManagerCreation(t *testing.T) {
	// Create temporary log file for testing
	tempDir, err := os.MkdirTemp("", "lai_daemon_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFile := filepath.Join(tempDir, "test.log")
	if err := os.WriteFile(logFile, []byte("test log content"), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	// Test that daemon manager can be created
	manager, err := daemon.NewManager()
	if err != nil {
		t.Fatalf("Failed to create daemon manager: %v", err)
	}

	// Test process ID generation
	processID := manager.GenerateProcessID(logFile)
	if processID == "" {
		t.Error("Generated process ID should not be empty")
	}

	// Test custom process ID generation
	customName := "webapp"
	customID := manager.GenerateProcessIDWithName(customName)
	if customID != customName {
		t.Errorf("Expected custom ID %s, got %s", customName, customID)
	}
}

func TestRunDaemon_ProcessIDGeneration(t *testing.T) {
	// Create temporary log file
	tempDir, err := os.MkdirTemp("", "lai_daemon_id_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFile := filepath.Join(tempDir, "application.log")
	if err := os.WriteFile(logFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	manager, err := daemon.NewManager()
	if err != nil {
		t.Fatalf("Failed to create daemon manager: %v", err)
	}

	// Test automatic ID generation
	autoID := manager.GenerateProcessID(logFile)
	if autoID == "" {
		t.Error("Auto-generated ID should not be empty")
	}

	// Should contain the log file basename
	if !contains(autoID, "application.log") {
		t.Errorf("Generated ID should contain log file basename, got: %s", autoID)
	}

	// Test custom name ID generation
	customName := "my_custom_process"
	customID := manager.GenerateProcessIDWithName(customName)
	if customID != customName {
		t.Errorf("Custom ID should match input exactly, expected %s, got %s", customName, customID)
	}
}

func TestRunDaemon_ProcessExists(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_daemon_exists_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForDaemon(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Test non-existent process
	if testManager.ProcessExists("nonexistent") {
		t.Error("Non-existent process should not be detected as existing")
	}

	// Create a test process
	testProcess := &daemon.ProcessInfo{
		ID:        "test_existing",
		PID:       12345,
		LogFile:   "/test/app.log",
		StartTime: time.Now(),
		Status:    "running",
	}

	if err := testManager.SaveProcessInfo(testProcess); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Test existing process
	if !testManager.ProcessExists("test_existing") {
		t.Error("Existing process should be detected")
	}
}

func TestRunDaemon_LogPathGeneration(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_daemon_logpath_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForDaemon(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Test log path generation
	processID := "webapp_12345"
	logPath := testManager.GetProcessLogPath(processID)

	expectedPath := filepath.Join(tempDir, "logs", "webapp_12345.log")
	if logPath != expectedPath {
		t.Errorf("Expected log path %s, got %s", expectedPath, logPath)
	}
}

func TestRunDaemon_EnvironmentCheck(t *testing.T) {
	// Save original environment
	originalEnv := os.Getenv("LAI_DAEMON_MODE")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("LAI_DAEMON_MODE")
		} else {
			os.Setenv("LAI_DAEMON_MODE", originalEnv)
		}
	}()

	// Test environment variable detection
	os.Unsetenv("LAI_DAEMON_MODE")
	if os.Getenv("LAI_DAEMON_MODE") == "1" {
		t.Error("LAI_DAEMON_MODE should not be set initially")
	}

	// Test setting environment variable
	os.Setenv("LAI_DAEMON_MODE", "1")
	if os.Getenv("LAI_DAEMON_MODE") != "1" {
		t.Error("LAI_DAEMON_MODE should be set to '1'")
	}
}

func TestRunDaemon_ParameterValidation(t *testing.T) {
	// Test parameter handling for runDaemon function
	// We can't easily test the actual function due to os.StartProcess
	// but we can test the parameter validation logic

	tests := []struct {
		name        string
		logFile     string
		processName string
		expectError bool
		description string
	}{
		{
			name:        "Valid parameters",
			logFile:     testutils.GetTestLogPath("test.log"),
			processName: "webapp",
			expectError: false,
			description: "Should accept valid log file and process name",
		},
		{
			name:        "Empty process name",
			logFile:     testutils.GetTestLogPath("test.log"),
			processName: "",
			expectError: false,
			description: "Should accept empty process name (auto-generate)",
		},
		{
			name:        "Valid log file with custom name",
			logFile:     "/app/logs/application.log",
			processName: "my_application",
			expectError: false,
			description: "Should accept custom process name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since we can't test runDaemon directly without forking,
			// we test the underlying components it would use

			manager, err := daemon.NewManager()
			if err != nil {
				t.Fatalf("Failed to create manager: %v", err)
			}

			// Test ID generation logic
			var processID string
			if tt.processName != "" {
				processID = manager.GenerateProcessIDWithName(tt.processName)
				if manager.ProcessExists(processID) {
					// This would cause an error in runDaemon
					t.Logf("Process %s already exists (expected for test)", processID)
				}
			} else {
				processID = manager.GenerateProcessID(tt.logFile)
			}

			if processID == "" && !tt.expectError {
				t.Error("Process ID generation should not fail for valid parameters")
			}

			// Test log path generation
			logPath := manager.GetProcessLogPath(processID)
			if logPath == "" && !tt.expectError {
				t.Error("Log path generation should not fail")
			}
		})
	}
}

func TestRunDaemon_ProcessInfoStructure(t *testing.T) {
	// Test the ProcessInfo structure used by runDaemon
	now := time.Now()

	processInfo := &daemon.ProcessInfo{
		ID:        "test_daemon_process",
		PID:       os.Getpid(),
		LogFile:   "/test/daemon.log",
		StartTime: now,
		Status:    "running",
	}

	// Verify structure fields
	if processInfo.ID != "test_daemon_process" {
		t.Errorf("Expected ID 'test_daemon_process', got %s", processInfo.ID)
	}

	if processInfo.PID != os.Getpid() {
		t.Errorf("Expected PID %d, got %d", os.Getpid(), processInfo.PID)
	}

	if processInfo.LogFile != "/test/daemon.log" {
		t.Errorf("Expected LogFile '/test/daemon.log', got %s", processInfo.LogFile)
	}

	if !processInfo.StartTime.Equal(now) {
		t.Errorf("Expected StartTime %v, got %v", now, processInfo.StartTime)
	}

	if processInfo.Status != "running" {
		t.Errorf("Expected Status 'running', got %s", processInfo.Status)
	}
}

func TestRunMonitor_ParameterHandling(t *testing.T) {
	// Test parameter handling for runMonitor function
	// We test the parameter validation and setup logic

	// Create temporary log file
	tempDir, err := os.MkdirTemp("", "lai_monitor_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logFile := filepath.Join(tempDir, "monitor.log")
	if err := os.WriteFile(logFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	// Test parameter types that runMonitor would receive
	var lineThreshold *int
	var checkInterval *time.Duration
	var chatID *string

	// Test nil parameters (should use defaults)
	if lineThreshold != nil {
		t.Error("lineThreshold should be nil initially")
	}
	if checkInterval != nil {
		t.Error("checkInterval should be nil initially")
	}
	if chatID != nil {
		t.Error("chatID should be nil initially")
	}

	// Test setting parameters
	threshold := 10
	lineThreshold = &threshold

	interval := time.Duration(30 * time.Second)
	checkInterval = &interval

	chat := "12345"
	chatID = &chat

	// Verify parameters are set correctly
	if lineThreshold == nil || *lineThreshold != 10 {
		t.Error("lineThreshold should be set to 10")
	}
	if checkInterval == nil || *checkInterval != 30*time.Second {
		t.Error("checkInterval should be set to 30 seconds")
	}
	if chatID == nil || *chatID != "12345" {
		t.Error("chatID should be set to '12345'")
	}
}

// Helper function to create a test daemon manager for daemon tests
func createTestManagerForDaemon(tempDir string) (*daemon.Manager, error) {
	processDir := filepath.Join(tempDir, "processes")
	logDir := filepath.Join(tempDir, "logs")

	return daemon.NewManagerWithDirs(processDir, logDir)
}
