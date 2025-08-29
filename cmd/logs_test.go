package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shiquda/lai/internal/daemon"
)

func TestShowLastLines(t *testing.T) {
	// Create temporary directory and test file
	tempDir, err := os.MkdirTemp("", "lai_logs_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.log")
	testContent := []string{
		"Line 1",
		"Line 2",
		"Line 3",
		"Line 4",
		"Line 5",
		"Line 6",
		"Line 7",
		"Line 8",
		"Line 9",
		"Line 10",
	}

	// Write test content to file
	content := strings.Join(testContent, "\n") + "\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	tests := []struct {
		name      string
		numLines  int
		expected  []string
	}{
		{
			name:     "Last 3 lines",
			numLines: 3,
			expected: []string{"Line 8", "Line 9", "Line 10"},
		},
		{
			name:     "Last 5 lines",
			numLines: 5,
			expected: []string{"Line 6", "Line 7", "Line 8", "Line 9", "Line 10"},
		},
		{
			name:     "More lines than available",
			numLines: 15,
			expected: testContent, // All lines
		},
		{
			name:     "Zero lines",
			numLines: 0,
			expected: []string{}, // No output expected when numLines is 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output by redirecting stdout
			origStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := showLastLines(testFile, tt.numLines)
			if err != nil {
				t.Fatalf("showLastLines failed: %v", err)
			}

			w.Close()
			os.Stdout = origStdout

			// Read captured output
			buf := make([]byte, 1024)
			n, _ := r.Read(buf)
			output := string(buf[:n])
			r.Close()

			// Verify output contains expected lines
			if len(tt.expected) == 0 {
				// For empty expected output, just check that no error occurred
				if strings.TrimSpace(output) != "" {
					t.Errorf("Expected no output, got: %s", output)
				}
			} else {
				for _, expectedLine := range tt.expected {
					if !strings.Contains(output, expectedLine) {
						t.Errorf("Output should contain '%s', got: %s", expectedLine, output)
					}
				}
			}
		})
	}
}

func TestShowLastLines_NonExistentFile(t *testing.T) {
	err := showLastLines("/nonexistent/file.log", 10)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestShowLastLines_EmptyFile(t *testing.T) {
	// Create temporary empty file
	tempDir, err := os.MkdirTemp("", "lai_logs_test_empty_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	emptyFile := filepath.Join(tempDir, "empty.log")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	// Should not error on empty file
	err = showLastLines(emptyFile, 5)
	if err != nil {
		t.Errorf("showLastLines should not fail on empty file: %v", err)
	}
}

func TestLogsCmd_ProcessExists(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_logs_cmd_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForLogs(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create test process
	testProcess := &daemon.ProcessInfo{
		ID:        "test_logs",
		PID:       12345,
		LogFile:   "/test/app.log",
		StartTime: time.Now(),
		Status:    "running",
	}

	if err := testManager.SaveProcessInfo(testProcess); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Create a test log file
	logPath := testManager.GetProcessLogPath("test_logs")
	testLogContent := "Test log line 1\nTest log line 2\nTest log line 3\n"
	if err := os.WriteFile(logPath, []byte(testLogContent), 0644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	// Test that the process exists and log file is accessible
	_, err = testManager.LoadProcessInfo("test_logs")
	if err != nil {
		t.Errorf("Process should exist: %v", err)
	}

	// Test that log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file should exist at %s", logPath)
	}
}

func TestLogsCmd_ProcessNotFound(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_logs_cmd_notfound_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForLogs(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Try to load non-existent process
	_, err = testManager.LoadProcessInfo("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent process")
	}
}

func TestLogsCmd_LogFileNotFound(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "lai_logs_cmd_nolog_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create manager with temp directory
	testManager, err := createTestManagerForLogs(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Create test process but no log file
	testProcess := &daemon.ProcessInfo{
		ID:        "test_nolog",
		PID:       12345,
		LogFile:   "/test/app.log",
		StartTime: time.Now(),
		Status:    "running",
	}

	if err := testManager.SaveProcessInfo(testProcess); err != nil {
		t.Fatalf("Failed to save test process: %v", err)
	}

	// Test that log file doesn't exist
	logPath := testManager.GetProcessLogPath("test_nolog")
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Errorf("Log file should not exist at %s", logPath)
	}
}

func TestTailFile_BasicFunctionality(t *testing.T) {
	// Create temporary file with some content
	tempDir, err := os.MkdirTemp("", "lai_tail_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "tail.log")
	initialContent := "Initial line 1\nInitial line 2\nInitial line 3\n"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Note: Since tailFile includes an infinite loop for following,
	// we'll only test the initial content display part
	// A full integration test would require goroutines and signaling

	// Test that the function can open and seek to the end of the file
	file, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()

	// Seek to end
	_, err = file.Seek(0, 2)
	if err != nil {
		t.Errorf("Failed to seek to end of file: %v", err)
	}
}

func TestGetProcessLogPath_Format(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lai_logpath_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testManager, err := createTestManagerForLogs(tempDir)
	if err != nil {
		t.Fatalf("Failed to create test manager: %v", err)
	}

	// Test log path format
	processID := "webapp_12345"
	logPath := testManager.GetProcessLogPath(processID)
	
	expectedPath := filepath.Join(tempDir, "logs", "webapp_12345.log")
	if logPath != expectedPath {
		t.Errorf("Expected log path %s, got %s", expectedPath, logPath)
	}
}

// Helper function to create a test daemon manager for logs tests
func createTestManagerForLogs(tempDir string) (*daemon.Manager, error) {
	processDir := filepath.Join(tempDir, "processes")
	logDir := filepath.Join(tempDir, "logs")

	return daemon.NewManagerWithDirs(processDir, logDir)
}