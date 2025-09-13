package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigValidate_WithValidPromptTemplates(t *testing.T) {
	cfg := &Config{
		LogFile:   "/path/to/log.log",
		OpenAI:    OpenAIConfig{APIKey: "test-key"},
		Language:  "English",
		Notifications: NotificationsConfig{
			Providers: map[string]ServiceConfig{
				"telegram": {
					Enabled:  true,
					Provider: "telegram",
					Config:   map[string]interface{}{"bot_token": "test", "chat_id": "123"},
				},
			},
		},
		PromptTemplates: PromptTemplatesConfig{
			SummarizeTemplate:        "Analyze {{log_content}} in {{language}}",
			ErrorAnalysisTemplate:    "Check {{log_content}} for errors in {{language}}",
			CustomVariables: map[string]string{
				"app_name": "TestApp",
			},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestConfigValidate_WithInvalidPromptTemplates(t *testing.T) {
	cfg := &Config{
		LogFile:   "/path/to/log.log",
		OpenAI:    OpenAIConfig{APIKey: "test-key"},
		Language:  "English",
		Notifications: NotificationsConfig{
			Providers: map[string]ServiceConfig{
				"telegram": {
					Enabled:  true,
					Provider: "telegram",
					Config:   map[string]interface{}{"bot_token": "test", "chat_id": "123"},
				},
			},
		},
		PromptTemplates: PromptTemplatesConfig{
			SummarizeTemplate: "Analyze {{unknown_variable}} in {{language}}",
			CustomVariables: map[string]string{
				"app_name": "TestApp",
			},
		},
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt template validation failed")
	assert.Contains(t, err.Error(), "undefined variables found")
}

func TestConfigValidate_WithEmptyPromptTemplates(t *testing.T) {
	cfg := &Config{
		LogFile:   "/path/to/log.log",
		OpenAI:    OpenAIConfig{APIKey: "test-key"},
		Language:  "English",
		Notifications: NotificationsConfig{
			Providers: map[string]ServiceConfig{
				"telegram": {
					Enabled:  true,
					Provider: "telegram",
					Config:   map[string]interface{}{"bot_token": "test", "chat_id": "123"},
				},
			},
		},
		PromptTemplates: PromptTemplatesConfig{
			SummarizeTemplate:     "",
			ErrorAnalysisTemplate: "",
			CustomVariables:      make(map[string]string),
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err) // Empty templates should be valid (use built-in)
}

func TestConfigValidate_WithCustomVariablesInTemplate(t *testing.T) {
	cfg := &Config{
		LogFile:   "/path/to/log.log",
		OpenAI:    OpenAIConfig{APIKey: "test-key"},
		Language:  "English",
		Notifications: NotificationsConfig{
			Providers: map[string]ServiceConfig{
				"telegram": {
					Enabled:  true,
					Provider: "telegram",
					Config:   map[string]interface{}{"bot_token": "test", "chat_id": "123"},
				},
			},
		},
		PromptTemplates: PromptTemplatesConfig{
			SummarizeTemplate: "Analyze {{log_content}} from {{app_name}} in {{environment}}",
			CustomVariables: map[string]string{
				"app_name":    "MyApp",
				"environment": "production",
			},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err) // Custom variables should be allowed
}

func TestConfigValidate_TemplateWithMultipleFormats(t *testing.T) {
	cfg := &Config{
		LogFile:   "/path/to/log.log",
		OpenAI:    OpenAIConfig{APIKey: "test-key"},
		Language:  "English",
		Notifications: NotificationsConfig{
			Providers: map[string]ServiceConfig{
				"telegram": {
					Enabled:  true,
					Provider: "telegram",
					Config:   map[string]interface{}{"bot_token": "test", "chat_id": "123"},
				},
			},
		},
		PromptTemplates: PromptTemplatesConfig{
			SummarizeTemplate: "Analyze {{log_content}} in ${language} for $app_name",
			CustomVariables: map[string]string{
				"app_name": "TestApp",
			},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err) // All formats should be supported
}

func TestConfigValidate_ErrorAnalysisTemplateOnly(t *testing.T) {
	cfg := &Config{
		LogFile:   "/path/to/log.log",
		OpenAI:    OpenAIConfig{APIKey: "test-key"},
		Language:  "English",
		Notifications: NotificationsConfig{
			Providers: map[string]ServiceConfig{
				"telegram": {
					Enabled:  true,
					Provider: "telegram",
					Config:   map[string]interface{}{"bot_token": "test", "chat_id": "123"},
				},
			},
		},
		PromptTemplates: PromptTemplatesConfig{
			SummarizeTemplate:        "",
			ErrorAnalysisTemplate:    "Check {{log_content}} for errors in {{language}}",
			CustomVariables: map[string]string{
				"app_name": "TestApp",
			},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err) // Only error analysis template should be valid
}

func TestConfigValidate_SummarizeTemplateOnly(t *testing.T) {
	cfg := &Config{
		LogFile:   "/path/to/log.log",
		OpenAI:    OpenAIConfig{APIKey: "test-key"},
		Language:  "English",
		Notifications: NotificationsConfig{
			Providers: map[string]ServiceConfig{
				"telegram": {
					Enabled:  true,
					Provider: "telegram",
					Config:   map[string]interface{}{"bot_token": "test", "chat_id": "123"},
				},
			},
		},
		PromptTemplates: PromptTemplatesConfig{
			SummarizeTemplate:     "Analyze {{log_content}} in {{language}}",
			ErrorAnalysisTemplate: "",
			CustomVariables:      map[string]string{},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err) // Only summarize template should be valid
}

func TestConfigValidate_TemplateWithBuiltinVariables(t *testing.T) {
	cfg := &Config{
		LogFile:   "/path/to/log.log",
		OpenAI:    OpenAIConfig{APIKey: "test-key"},
		Language:  "English",
		Notifications: NotificationsConfig{
			Providers: map[string]ServiceConfig{
				"telegram": {
					Enabled:  true,
					Provider: "telegram",
					Config:   map[string]interface{}{"bot_token": "test", "chat_id": "123"},
				},
			},
		},
		PromptTemplates: PromptTemplatesConfig{
			SummarizeTemplate: "Analyze {{log_content}} for {{system}} version {{version}} at {{timestamp}}",
			CustomVariables: map[string]string{},
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err) // Built-in variables should be allowed
}

func TestValidatePromptTemplates_InvalidTemplate(t *testing.T) {
	cfg := &Config{
		PromptTemplates: PromptTemplatesConfig{
			SummarizeTemplate: "Analyze {{log_content}} with {{invalid_var}}",
			CustomVariables: map[string]string{},
		},
	}

	err := cfg.validatePromptTemplates()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undefined variables found")
}

func TestValidatePromptTemplates_EmptyTemplate(t *testing.T) {
	cfg := &Config{
		PromptTemplates: PromptTemplatesConfig{
			SummarizeTemplate:     "",
			ErrorAnalysisTemplate: "",
			CustomVariables:      map[string]string{},
		},
	}

	err := cfg.validatePromptTemplates()
	assert.NoError(t, err) // Empty templates should not cause validation errors
}

func TestValidatePromptTemplates_ValidWithCustomVariables(t *testing.T) {
	cfg := &Config{
		PromptTemplates: PromptTemplatesConfig{
			SummarizeTemplate: "Analyze {{log_content}} from {{app_name}}",
			CustomVariables: map[string]string{
				"app_name": "TestApp",
			},
		},
	}

	err := cfg.validatePromptTemplates()
	assert.NoError(t, err)
}