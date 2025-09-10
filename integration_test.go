package main

import (
	"os"
	"testing"
	"time"

	"github.com/shiquda/lai/internal/collector"
	"github.com/shiquda/lai/internal/config"
	"github.com/shiquda/lai/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestCollectorWorkflow(t *testing.T) {
	// Create temp log file
	tempDir, cleanup := testutils.CreateTempDir(t)
	defer cleanup()

	logFile := testutils.CreateFileWithContent(t, tempDir, "test.log", "Initial line\n")

	// Set up collector
	logCollector := collector.New(logFile, 2, 50*time.Millisecond)

	// Track trigger calls
	var triggerContent string
	var triggerCount int

	logCollector.SetTriggerHandler(func(newContent string) error {
		triggerCount++
		triggerContent = newContent
		return nil
	})

	// Start collector
	done := make(chan bool)
	go func() {
		logCollector.Start()
		done <- true
	}()

	// Wait for initialization
	time.Sleep(100 * time.Millisecond)

	// Add new lines to trigger handler
	testutils.AppendToFile(t, logFile, "New line 1\nNew line 2\n")

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Verify trigger was called
	assert.Equal(t, 1, triggerCount, "Handler should be triggered once")
	assert.Contains(t, triggerContent, "New line 1")
	assert.Contains(t, triggerContent, "New line 2")
}

func TestConfigValidation(t *testing.T) {
	// Test valid config
	testLogPath := testutils.GetTestLogPath("test.log")
	validConfig := &config.Config{
		LogFile: testLogPath,
		OpenAI: config.OpenAIConfig{
			APIKey: "sk-test-123",
		},
		Notifications: config.NotificationsConfig{
			Providers: map[string]config.ServiceConfig{
				"telegram": {
					Enabled:  true,
					Provider: "telegram",
					Config: map[string]interface{}{
						"token":   "123:token",
						"chat_id": "-100123456789",
					},
				},
			},
		},
	}

	err := validConfig.Validate()
	assert.NoError(t, err)

	// Test invalid config - missing API key
	invalidConfig := &config.Config{
		LogFile: testLogPath,
		OpenAI: config.OpenAIConfig{
			APIKey: "", // Missing
		},
		Notifications: config.NotificationsConfig{
			Providers: map[string]config.ServiceConfig{
				"telegram": {
					Enabled:  true,
					Provider: "telegram",
					Config: map[string]interface{}{
						"token":   "123:token",
						"chat_id": "-100123456789",
					},
				},
			},
		},
	}

	err = invalidConfig.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "openai.api_key is required")
}

func TestConfigMerging(t *testing.T) {
	// Test that command line parameters override global config
	tempDir, cleanup := testutils.CreateTempDir(t)
	defer cleanup()

	// Set up temporary home directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create global config
	globalConfig := &config.GlobalConfig{
		Notifications: config.NotificationsConfig{
			OpenAI: config.OpenAIConfig{
				APIKey:  "global-key",
				BaseURL: "https://api.openai.com/v1",
				Model:   "gpt-3.5-turbo",
			},
			Providers: map[string]config.ServiceConfig{
				"telegram": {
					Enabled:  true,
					Provider: "telegram",
					Config: map[string]interface{}{
						"token":   "global:token",
						"chat_id": "-100global",
					},
				},
			},
		},
		Defaults: config.DefaultsConfig{
			LineThreshold: 15,
			CheckInterval: 60 * time.Second,
		},
	}

	err := config.SaveGlobalConfig(globalConfig)
	assert.NoError(t, err)

	// Test with overrides
	lineThreshold := 25
	checkInterval := 30 * time.Second
	testLogPath := testutils.GetTestLogPath("test.log")

	runtimeConfig, err := config.BuildRuntimeConfig(testLogPath, &lineThreshold, &checkInterval, nil)

	assert.NoError(t, err)
	assert.Equal(t, testLogPath, runtimeConfig.LogFile)
	assert.Equal(t, 25, runtimeConfig.LineThreshold)             // Overridden
	assert.Equal(t, 30*time.Second, runtimeConfig.CheckInterval) // Overridden
	// Check if telegram provider is configured with correct chat_id
	if telegramProvider, exists := runtimeConfig.Notifications.Providers["telegram"]; exists {
		assert.Equal(t, "-100global", telegramProvider.Config["chat_id"])
	}
	assert.Equal(t, "global-key", runtimeConfig.OpenAI.APIKey) // From global
}
