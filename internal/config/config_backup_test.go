package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create a test config file
	testConfig := `version: "1.0.0"
notifications:
  openai:
    api_key: "test-key"
defaults:
  line_threshold: 10
logging:
  level: "info"`

	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Test backup creation
	backupPath, err := BackupConfig(configPath)
	if err != nil {
		t.Fatalf("BackupConfig failed: %v", err)
	}

	// Verify backup file exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup file was not created: %s", backupPath)
	}

	// Verify backup filename contains version and timestamp
	if !contains(backupPath, "v1.0.0") {
		t.Errorf("Backup filename should contain version: %s", backupPath)
	}

	// Verify backup content
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup content: %v", err)
	}

	if string(backupContent) != testConfig {
		t.Errorf("Backup content mismatch.\nExpected: %s\nGot: %s", testConfig, string(backupContent))
	}
}

func TestBackupConfigWithUnknownVersion(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create a test config file without version
	testConfig := `notifications:
  openai:
    api_key: "test-key"
defaults:
  line_threshold: 10`

	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Test backup creation
	backupPath, err := BackupConfig(configPath)
	if err != nil {
		t.Fatalf("BackupConfig failed: %v", err)
	}

	// Verify backup filename contains "unknown" version
	if !contains(backupPath, "vunknown") {
		t.Errorf("Backup filename should contain 'unknown' version: %s", backupPath)
	}
}

func TestBackupConfigNonExistent(t *testing.T) {
	// Test backup creation for non-existent config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "nonexistent.yaml")

	_, err := BackupConfig(configPath)
	if err == nil {
		t.Error("Expected error for non-existent config, got nil")
	}
}

func TestRestoreConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	backupPath := filepath.Join(tempDir, "config.v1.0.0.20240101-120000.backup.yaml")

	// Create a backup file
	backupContent := `version: "1.0.0"
notifications:
  openai:
    api_key: "restored-key"
    model: "gpt-4"
defaults:
  line_threshold: 15
  language: "Chinese"`

	err := os.WriteFile(backupPath, []byte(backupContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	// Test restore
	err = RestoreConfig(backupPath, configPath)
	if err != nil {
		t.Fatalf("RestoreConfig failed: %v", err)
	}

	// Verify config file was restored
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not restored: %s", configPath)
	}

	// Verify restored content
	configContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read restored config: %v", err)
	}

	if string(configContent) != backupContent {
		t.Errorf("Restored content mismatch.\nExpected: %s\nGot: %s", backupContent, string(configContent))
	}
}

func TestRestoreConfigNonExistentBackup(t *testing.T) {
	// Test restore with non-existent backup
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	backupPath := filepath.Join(tempDir, "nonexistent.backup.yaml")

	err := RestoreConfig(backupPath, configPath)
	if err == nil {
		t.Error("Expected error for non-existent backup, got nil")
	}
}

func TestRestoreConfigInvalidBackup(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	backupPath := filepath.Join(tempDir, "invalid.backup.yaml")

	// Create an invalid backup file
	invalidContent := `this is not a valid yaml config`
	err := os.WriteFile(backupPath, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid backup: %v", err)
	}

	// Test restore should fail
	err = RestoreConfig(backupPath, configPath)
	if err == nil {
		t.Error("Expected error for invalid backup, got nil")
	}
}

func TestListBackups(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create some backup files
	backups := []string{
		"config.v1.0.0.20240101-120000.backup.yaml",
		"config.v1.0.1.20240102-120000.backup.yaml",
		"config.vdev.20240103-120000.backup.yaml",
		"not-a-backup.txt", // This should be ignored
	}

	for _, backup := range backups {
		path := filepath.Join(tempDir, backup)
		err := os.WriteFile(path, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create backup file %s: %v", backup, err)
		}
	}

	// Test listing backups
	listedBackups, err := ListBackups(configPath)
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	// Should only list 3 backup files (ignore non-backup files)
	if len(listedBackups) != 3 {
		t.Errorf("Expected 3 backups, got %d: %v", len(listedBackups), listedBackups)
	}

	// Verify all expected backups are listed
	for _, expected := range backups[:3] {
		found := false
		for _, listed := range listedBackups {
			if contains(listed, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected backup %s not found in list", expected)
		}
	}
}

func TestCleanupOldBackups(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create several backup files with different timestamps
	backups := []string{
		"config.v1.0.0.20240101-120000.backup.yaml", // Oldest
		"config.v1.0.1.20240102-120000.backup.yaml",
		"config.v1.0.2.20240103-120000.backup.yaml", // Newest
	}

	for _, backup := range backups {
		path := filepath.Join(tempDir, backup)
		err := os.WriteFile(path, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create backup file %s: %v", backup, err)
		}
	}

	// Test cleanup, keeping only 2 most recent
	err := CleanupOldBackups(configPath, 2)
	if err != nil {
		t.Fatalf("CleanupOldBackups failed: %v", err)
	}

	// Should have only 2 backups remaining
	remainingBackups, err := ListBackups(configPath)
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	if len(remainingBackups) != 2 {
		t.Errorf("Expected 2 backups after cleanup, got %d", len(remainingBackups))
	}

	// Verify the newest backups are kept
	for _, remaining := range remainingBackups {
		if contains(remaining, "20240101-120000") {
			t.Error("Oldest backup should have been removed")
		}
	}
}

func TestCleanupOldBackupsNoCleanupNeeded(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create only 1 backup file
	backupPath := filepath.Join(tempDir, "config.v1.0.0.20240101-120000.backup.yaml")
	err := os.WriteFile(backupPath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	// Test cleanup, keeping 3 backups (should not remove anything)
	err = CleanupOldBackups(configPath, 3)
	if err != nil {
		t.Fatalf("CleanupOldBackups failed: %v", err)
	}

	// Should still have 1 backup
	remainingBackups, err := ListBackups(configPath)
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	if len(remainingBackups) != 1 {
		t.Errorf("Expected 1 backup after cleanup, got %d", len(remainingBackups))
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
			 s[len(s)-len(substr):] == substr ||
			 func() bool {
					for i := 0; i <= len(s)-len(substr); i++ {
						if s[i:i+len(substr)] == substr {
							return true
						}
					}
					return false
				}())))
}