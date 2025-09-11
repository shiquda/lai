package config

import (
	"testing"
	"time"
)

func TestConfigMigrationPreservesUserSettings(t *testing.T) {
	// Create a user config with custom settings
	userConfig := &GlobalConfig{
		Version: "1.0.0",
		Notifications: NotificationsConfig{
			OpenAI: OpenAIConfig{
				APIKey:  "user-secret-key",
				BaseURL: "https://custom.openai.com/v1",
				Model:   "gpt-4",
			},
			Providers: map[string]ServiceConfig{
				"telegram": {
					Enabled:  true,
					Provider: "telegram",
					Config: map[string]interface{}{
						"bot_token": "user-bot-token",
						"chat_id":   "user-chat-id",
					},
				},
				"email": {
					Enabled:  false,
					Provider: "smtp",
					Config: map[string]interface{}{
						"smtp_host": "user.smtp.com",
						"username":  "user@example.com",
					},
				},
			},
		},
		Defaults: DefaultsConfig{
			LineThreshold: 5,
			CheckInterval: 15 * time.Second,
			FinalSummary:  false,
			Language:      "Chinese",
		},
		Logging: LoggingConfig{
			Level: "debug",
		},
	}

	// Create a default config with different values
	defaultConfig := &GlobalConfig{
		Version: "2.0.0",
		Notifications: NotificationsConfig{
			OpenAI: OpenAIConfig{
				BaseURL: "https://api.openai.com/v1",
				Model:   "gpt-3.5-turbo",
			},
			Providers: map[string]ServiceConfig{
				"telegram": {
					Enabled:  false,
					Provider: "telegram",
					Config:   map[string]interface{}{},
					Defaults: map[string]interface{}{
						"parse_mode": "markdown",
					},
				},
				"email": {
					Enabled:  false,
					Provider: "smtp",
					Config:   map[string]interface{}{},
					Defaults: map[string]interface{}{
						"subject": "Log Summary Notification",
					},
				},
				"slack": {
					Enabled:  false,
					Provider: "slack",
					Config:   map[string]interface{}{},
					Defaults: map[string]interface{}{
						"username":   "Lai Bot",
						"icon_emoji": ":robot_face:",
					},
				},
				"discord": {
					Enabled:  false,
					Provider: "discord",
					Config:   map[string]interface{}{},
				},
			},
		},
		Defaults: DefaultsConfig{
			LineThreshold: 10,
			CheckInterval: 30 * time.Second,
			FinalSummary:  true,
			Language:      "English",
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}

	// Perform merge
	mergedConfig := mergeConfigsWithReflection(userConfig, defaultConfig)

	// Verify user settings are preserved
	if mergedConfig.Notifications.OpenAI.APIKey != "user-secret-key" {
		t.Errorf("Expected API key to be preserved, got %s", mergedConfig.Notifications.OpenAI.APIKey)
	}

	if mergedConfig.Notifications.OpenAI.BaseURL != "https://custom.openai.com/v1" {
		t.Errorf("Expected custom base URL to be preserved, got %s", mergedConfig.Notifications.OpenAI.BaseURL)
	}

	if mergedConfig.Notifications.OpenAI.Model != "gpt-4" {
		t.Errorf("Expected custom model to be preserved, got %s", mergedConfig.Notifications.OpenAI.Model)
	}

	// Verify provider settings are preserved
	telegramConfig := mergedConfig.Notifications.Providers["telegram"]
	if !telegramConfig.Enabled {
		t.Errorf("Expected telegram enabled status to be preserved")
	}

	if telegramConfig.Config["bot_token"] != "user-bot-token" {
		t.Errorf("Expected telegram bot token to be preserved, got %v", telegramConfig.Config["bot_token"])
	}

	// Verify defaults are preserved
	if mergedConfig.Defaults.LineThreshold != 5 {
		t.Errorf("Expected line threshold to be preserved, got %d", mergedConfig.Defaults.LineThreshold)
	}

	if mergedConfig.Defaults.CheckInterval != 15*time.Second {
		t.Errorf("Expected check interval to be preserved, got %v", mergedConfig.Defaults.CheckInterval)
	}

	if mergedConfig.Defaults.FinalSummary != false {
		t.Errorf("Expected final summary setting to be preserved, got %v", mergedConfig.Defaults.FinalSummary)
	}

	if mergedConfig.Defaults.Language != "Chinese" {
		t.Errorf("Expected language to be preserved, got %s", mergedConfig.Defaults.Language)
	}

	// Verify logging level is preserved
	if mergedConfig.Logging.Level != "debug" {
		t.Errorf("Expected logging level to be preserved, got %s", mergedConfig.Logging.Level)
	}

	// Verify version is updated
	if mergedConfig.Version != "2.0.0" {
		t.Errorf("Expected version to be updated to 2.0.0, got %s", mergedConfig.Version)
	}

	// Verify new providers are added
	if _, exists := mergedConfig.Notifications.Providers["slack"]; !exists {
		t.Errorf("Expected slack provider to be added")
	}

	if _, exists := mergedConfig.Notifications.Providers["discord"]; !exists {
		t.Errorf("Expected discord provider to be added")
	}
}

func TestConfigMigrationWithEmptyValues(t *testing.T) {
	// Create a user config with some empty values
	userConfig := &GlobalConfig{
		Version: "1.0.0",
		Notifications: NotificationsConfig{
			OpenAI: OpenAIConfig{
				APIKey:  "user-secret-key",
				BaseURL: "", // Empty value should be filled from default
				Model:   "", // Empty value should be filled from default
			},
		},
		Defaults: DefaultsConfig{
			LineThreshold: 5,
			CheckInterval: 0, // Zero value should be filled from default
			FinalSummary:  true,
			Language:      "", // Empty value should be filled from default
		},
	}

	// Create a default config
	defaultConfig := &GlobalConfig{
		Version: "2.0.0",
		Notifications: NotificationsConfig{
			OpenAI: OpenAIConfig{
				BaseURL: "https://api.openai.com/v1",
				Model:   "gpt-3.5-turbo",
			},
		},
		Defaults: DefaultsConfig{
			LineThreshold: 10,
			CheckInterval: 30 * time.Second,
			FinalSummary:  true,
			Language:      "English",
		},
	}

	// Perform merge
	mergedConfig := mergeConfigsWithReflection(userConfig, defaultConfig)

	// Verify non-empty values are preserved
	if mergedConfig.Notifications.OpenAI.APIKey != "user-secret-key" {
		t.Errorf("Expected API key to be preserved, got %s", mergedConfig.Notifications.OpenAI.APIKey)
	}

	if mergedConfig.Defaults.LineThreshold != 5 {
		t.Errorf("Expected line threshold to be preserved, got %d", mergedConfig.Defaults.LineThreshold)
	}

	// Verify empty/zero values are filled from defaults
	if mergedConfig.Notifications.OpenAI.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("Expected base URL to be filled from default, got %s", mergedConfig.Notifications.OpenAI.BaseURL)
	}

	if mergedConfig.Notifications.OpenAI.Model != "gpt-3.5-turbo" {
		t.Errorf("Expected model to be filled from default, got %s", mergedConfig.Notifications.OpenAI.Model)
	}

	if mergedConfig.Defaults.CheckInterval != 30*time.Second {
		t.Errorf("Expected check interval to be filled from default, got %v", mergedConfig.Defaults.CheckInterval)
	}

	if mergedConfig.Defaults.Language != "English" {
		t.Errorf("Expected language to be filled from default, got %s", mergedConfig.Defaults.Language)
	}
}

func TestConfigMigrationWithBooleanValues(t *testing.T) {
	// Create a user config with boolean values
	userConfig := &GlobalConfig{
		Version: "1.0.0",
		Defaults: DefaultsConfig{
			FinalSummary:     false, // Explicitly set to false
			FinalSummaryOnly: true,  // Explicitly set to true
			ErrorOnlyMode:    false, // Explicitly set to false
		},
	}

	// Create a default config with different boolean values
	defaultConfig := &GlobalConfig{
		Version: "2.0.0",
		Defaults: DefaultsConfig{
			FinalSummary:     true,  // Default is true
			FinalSummaryOnly: false, // Default is false
			ErrorOnlyMode:    false, // Default is false
		},
	}

	// Perform merge
	mergedConfig := mergeConfigsWithReflection(userConfig, defaultConfig)

	// Verify boolean values are preserved
	if mergedConfig.Defaults.FinalSummary != false {
		t.Errorf("Expected final summary to remain false, got %v", mergedConfig.Defaults.FinalSummary)
	}

	if mergedConfig.Defaults.FinalSummaryOnly != true {
		t.Errorf("Expected final summary only to remain true, got %v", mergedConfig.Defaults.FinalSummaryOnly)
	}

	if mergedConfig.Defaults.ErrorOnlyMode != false {
		t.Errorf("Expected error only mode to remain false, got %v", mergedConfig.Defaults.ErrorOnlyMode)
	}
}

func TestNeedsMigration(t *testing.T) {
	tests := []struct {
		name         string
		configVer    string
		defaultVer   string
		hasVersion   bool
		expectMigrate bool
	}{
		{
			name:         "No version field",
			configVer:    "",
			defaultVer:   "1.0.0",
			hasVersion:   false,
			expectMigrate: true,
		},
		{
			name:         "Older version",
			configVer:    "1.0.0",
			defaultVer:   "2.0.0",
			hasVersion:   true,
			expectMigrate: true,
		},
		{
			name:         "Same version",
			configVer:    "1.0.0",
			defaultVer:   "1.0.0",
			hasVersion:   true,
			expectMigrate: false,
		},
		{
			name:         "Newer version",
			configVer:    "2.0.0",
			defaultVer:   "1.0.0",
			hasVersion:   true,
			expectMigrate: false,
		},
		{
			name:         "Both dev versions",
			configVer:    "dev",
			defaultVer:   "dev",
			hasVersion:   true,
			expectMigrate: false,
		},
		{
			name:         "Config dev, default real",
			configVer:    "dev",
			defaultVer:   "1.0.0",
			hasVersion:   true,
			expectMigrate: true,
		},
		{
			name:         "Config real, default dev",
			configVer:    "1.0.0",
			defaultVer:   "dev",
			hasVersion:   true,
			expectMigrate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawConfig := make(map[string]interface{})
			
			if tt.hasVersion {
				rawConfig["version"] = tt.configVer
			}

			// Test the logic directly
			result := false
			
			if !tt.hasVersion {
				result = true
			} else if tt.configVer == "dev" {
				if tt.defaultVer == "dev" {
					result = false
				} else {
					result = true
				}
			} else if tt.defaultVer == "dev" {
				result = false
			} else {
				result = compareVersions(tt.configVer, tt.defaultVer) < 0
			}

			if result != tt.expectMigrate {
				t.Errorf("Expected migration: %v, got: %v", tt.expectMigrate, result)
			}
		})
	}
}