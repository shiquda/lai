package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// FieldType represents the type of a configuration field
type FieldType string

const (
	TypeString     FieldType = "string"
	TypeInt        FieldType = "int"
	TypeBool       FieldType = "bool"
	TypeDuration   FieldType = "duration"
	TypeStringList FieldType = "string_list"
	TypeSecret     FieldType = "secret" // For sensitive data like API keys
)

// Category represents configuration categories for organization
type Category string

const (
	CategoryGeneral       Category = "general"
	CategoryNotifications Category = "notifications"
	CategoryOpenAI        Category = "openai"
	CategoryProviders     Category = "providers"
	CategoryDefaults      Category = "defaults"
	CategoryLogging       Category = "logging"
)

// FieldMetadata describes metadata for a single configuration field
type FieldMetadata struct {
	Key          string    `json:"key"`          // Configuration key path (e.g., "notifications.openai.api_key")
	DisplayName  string    `json:"display_name"` // Human-readable name
	Description  string    `json:"description"`  // Detailed description
	Type         FieldType `json:"type"`         // Field type
	Category     Category  `json:"category"`     // Configuration category
	Required     bool      `json:"required"`     // Whether this field is required
	DefaultValue string    `json:"default"`      // Default value as string
	Validation   string    `json:"validation"`   // Validation pattern or rule
	Sensitive    bool      `json:"sensitive"`    // Whether to hide the value in display
	Examples     []string  `json:"examples"`     // Example values
	Level        int       `json:"level"`        // Nesting level for UI display (0=root, 1=child, etc.)
}

// ConfigSection represents a section of related configuration fields
type ConfigSection struct {
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	Description string          `json:"description"`
	Category    Category        `json:"category"`
	Fields      []FieldMetadata `json:"fields"`
	Level       int             `json:"level"`
}

// ConfigMetadata contains all configuration metadata
type ConfigMetadata struct {
	Sections []ConfigSection `json:"sections"`
}

// GetConfigMetadata returns the complete configuration metadata
func GetConfigMetadata() *ConfigMetadata {
	return &ConfigMetadata{
		Sections: []ConfigSection{
			{
				Name:        "general",
				DisplayName: "General Settings",
				Description: "Basic application configuration options",
				Category:    CategoryGeneral,
				Level:       0,
				Fields: []FieldMetadata{
					{
						Key:         "version",
						DisplayName: "Version",
						Description: "Configuration file version for compatibility management",
						Type:        TypeString,
						Category:    CategoryGeneral,
						Required:    false,
						Sensitive:   false,
						Level:       1,
					},
				},
			},
			{
				Name:        "defaults",
				DisplayName: "Default Configuration",
				Description: "Default application behavior settings",
				Category:    CategoryDefaults,
				Level:       0,
				Fields: []FieldMetadata{
					{
						Key:          "defaults.line_threshold",
						DisplayName:  "Line Threshold",
						Description:  "Number of new lines to trigger log summary generation",
						Type:         TypeInt,
						Category:     CategoryDefaults,
						Required:     false,
						DefaultValue: "10",
						Examples:     []string{"5", "10", "20"},
						Level:        1,
					},
					{
						Key:          "defaults.check_interval",
						DisplayName:  "Check Interval",
						Description:  "Time interval to check for log file changes",
						Type:         TypeDuration,
						Category:     CategoryDefaults,
						Required:     false,
						DefaultValue: "30s",
						Examples:     []string{"10s", "30s", "1m", "5m"},
						Level:        1,
					},
					{
						Key:          "defaults.language",
						DisplayName:  "AI Response Language",
						Description:  "Language used by AI when generating summaries",
						Type:         TypeString,
						Category:     CategoryDefaults,
						Required:     false,
						DefaultValue: "English",
						Examples:     []string{"English", "Chinese", "Japanese"},
						Level:        1,
					},
					{
						Key:          "defaults.final_summary",
						DisplayName:  "Final Summary",
						Description:  "Whether to send final summary when monitoring ends",
						Type:         TypeBool,
						Category:     CategoryDefaults,
						Required:     false,
						DefaultValue: "true",
						Level:        1,
					},
					{
						Key:          "defaults.final_summary_only",
						DisplayName:  "Final Summary Only",
						Description:  "Whether to send only final summary, skip intermediate summaries",
						Type:         TypeBool,
						Category:     CategoryDefaults,
						Required:     false,
						DefaultValue: "false",
						Level:        1,
					},
					{
						Key:          "defaults.error_only_mode",
						DisplayName:  "Error Only Mode",
						Description:  "Whether to send notifications only when errors are detected",
						Type:         TypeBool,
						Category:     CategoryDefaults,
						Required:     false,
						DefaultValue: "false",
						Level:        1,
					},
				},
			},
			{
				Name:        "openai",
				DisplayName: "OpenAI Configuration",
				Description: "OpenAI API related configuration for log summary generation",
				Category:    CategoryOpenAI,
				Level:       0,
				Fields: []FieldMetadata{
					{
						Key:         "notifications.openai.api_key",
						DisplayName: "API Key",
						Description: "OpenAI API key for calling GPT services",
						Type:        TypeSecret,
						Category:    CategoryOpenAI,
						Required:    true,
						Sensitive:   true,
						Examples:    []string{"sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},
						Level:       1,
					},
					{
						Key:          "notifications.openai.base_url",
						DisplayName:  "API Base URL",
						Description:  "Base URL for OpenAI API, can be used for proxy or custom endpoints",
						Type:         TypeString,
						Category:     CategoryOpenAI,
						Required:     false,
						DefaultValue: "https://api.openai.com/v1",
						Examples:     []string{"https://api.openai.com/v1", "https://your-proxy.com/v1"},
						Level:        1,
					},
					{
						Key:          "notifications.openai.model",
						DisplayName:  "Model Name",
						Description:  "OpenAI model to use",
						Type:         TypeString,
						Category:     CategoryOpenAI,
						Required:     false,
						DefaultValue: "gpt-3.5-turbo",
						Examples:     []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"},
						Level:        1,
					},
				},
			},
			{
				Name:        "telegram",
				DisplayName: "Telegram Notifications",
				Description: "Telegram bot notification configuration",
				Category:    CategoryProviders,
				Level:       0,
				Fields: []FieldMetadata{
					{
						Key:          "notifications.providers.telegram.enabled",
						DisplayName:  "Enable Telegram",
						Description:  "Whether to enable Telegram notifications",
						Type:         TypeBool,
						Category:     CategoryProviders,
						Required:     false,
						DefaultValue: "false",
						Level:        1,
					},
					{
						Key:         "notifications.providers.telegram.config.bot_token",
						DisplayName: "Bot Token",
						Description: "Telegram bot API token",
						Type:        TypeSecret,
						Category:    CategoryProviders,
						Required:    false,
						Sensitive:   true,
						Examples:    []string{"123456789:ABCdefGhIJKlmNOpqRsTUVwxyZ"},
						Level:       2,
					},
					{
						Key:         "notifications.providers.telegram.config.chat_id",
						DisplayName: "Chat ID",
						Description: "Telegram chat ID to receive messages",
						Type:        TypeString,
						Category:    CategoryProviders,
						Required:    false,
						Examples:    []string{"-1001234567890", "123456789"},
						Level:       2,
					},
				},
			},
			{
				Name:        "email",
				DisplayName: "Email Notifications",
				Description: "SMTP email notification configuration",
				Category:    CategoryProviders,
				Level:       0,
				Fields: []FieldMetadata{
					{
						Key:          "notifications.providers.email.enabled",
						DisplayName:  "Enable Email",
						Description:  "Whether to enable email notifications",
						Type:         TypeBool,
						Category:     CategoryProviders,
						Required:     false,
						DefaultValue: "false",
						Level:        1,
					},
					{
						Key:         "notifications.providers.email.config.smtp_host",
						DisplayName: "SMTP Server",
						Description: "SMTP server address",
						Type:        TypeString,
						Category:    CategoryProviders,
						Required:    false,
						Examples:    []string{"smtp.gmail.com", "smtp.qq.com", "smtp.163.com"},
						Level:       2,
					},
					{
						Key:          "notifications.providers.email.config.smtp_port",
						DisplayName:  "SMTP Port",
						Description:  "SMTP server port number",
						Type:         TypeInt,
						Category:     CategoryProviders,
						Required:     false,
						DefaultValue: "587",
						Examples:     []string{"25", "465", "587"},
						Level:        2,
					},
					{
						Key:         "notifications.providers.email.config.username",
						DisplayName: "Username",
						Description: "SMTP login username",
						Type:        TypeString,
						Category:    CategoryProviders,
						Required:    false,
						Level:       2,
					},
					{
						Key:         "notifications.providers.email.config.password",
						DisplayName: "Password",
						Description: "SMTP login password",
						Type:        TypeSecret,
						Category:    CategoryProviders,
						Required:    false,
						Sensitive:   true,
						Level:       2,
					},
					{
						Key:         "notifications.providers.email.config.from_email",
						DisplayName: "From Email",
						Description: "Email address to send from",
						Type:        TypeString,
						Category:    CategoryProviders,
						Required:    false,
						Examples:    []string{"noreply@example.com"},
						Level:       2,
					},
					{
						Key:         "notifications.providers.email.config.to_emails",
						DisplayName: "To Email List",
						Description: "List of email addresses to receive notifications",
						Type:        TypeStringList,
						Category:    CategoryProviders,
						Required:    false,
						Examples:    []string{"admin@example.com,user@example.com"},
						Level:       2,
					},
					{
						Key:          "notifications.providers.email.config.use_tls",
						DisplayName:  "Use TLS",
						Description:  "Whether to use TLS encryption for connection",
						Type:         TypeBool,
						Category:     CategoryProviders,
						Required:     false,
						DefaultValue: "true",
						Level:        2,
					},
				},
			},
			{
				Name:        "discord",
				DisplayName: "Discord Notifications",
				Description: "Discord notification configuration supporting bot tokens or webhooks",
				Category:    CategoryProviders,
				Level:       0,
				Fields: []FieldMetadata{
					{
						Key:          "notifications.providers.discord.enabled",
						DisplayName:  "Enable Discord",
						Description:  "Whether to enable Discord notifications",
						Type:         TypeBool,
						Category:     CategoryProviders,
						Required:     false,
						DefaultValue: "false",
						Level:        1,
					},
					{
						Key:          "notifications.providers.discord.provider",
						DisplayName:  "Provider Mode",
						Description:  "Select 'discord' for bot token integration or 'discord_webhook' for webhook delivery",
						Type:         TypeString,
						Category:     CategoryProviders,
						Required:     false,
						DefaultValue: "discord",
						Examples:     []string{"discord", "discord_webhook"},
						Level:        1,
					},
					{
						Key:         "notifications.providers.discord.config.bot_token",
						DisplayName: "Bot Token",
						Description: "Discord bot token used when provider is set to 'discord'",
						Type:        TypeSecret,
						Category:    CategoryProviders,
						Required:    false,
						Sensitive:   true,
						Examples:    []string{"MTIzNDU2Nzg5MA.ExAmPle.discord_bot_token_here"},
						Validation:  "^[A-Za-z0-9_-]+\\.[A-Za-z0-9_-]+\\.[A-Za-z0-9_-]+$",
						Level:       2,
					},
					{
						Key:         "notifications.providers.discord.config.channel_ids",
						DisplayName: "Channel IDs",
						Description: "Comma-separated list of Discord channel IDs used with the bot token provider",
						Type:        TypeStringList,
						Category:    CategoryProviders,
						Required:    false,
						Examples:    []string{"123456789012345678,234567890123456789"},
						Level:       2,
					},
					{
						Key:         "notifications.providers.discord.config.webhook_url",
						DisplayName: "Webhook URL",
						Description: "Discord channel Webhook URL used when provider is set to 'discord_webhook'",
						Type:        TypeSecret,
						Category:    CategoryProviders,
						Required:    false,
						Sensitive:   true,
						Examples:    []string{"https://discord.com/api/webhooks/123456789012345678/abcdefg..."},
						Validation:  "^https://discord\\.com/api/webhooks/\\d+/.+",
						Level:       2,
					},
					{
						Key:          "notifications.providers.discord.defaults.username",
						DisplayName:  "Default Username",
						Description:  "Username presented for Discord messages",
						Type:         TypeString,
						Category:     CategoryProviders,
						Required:     false,
						DefaultValue: "Lai Bot",
						Examples:     []string{"Lai Bot", "Log Monitor", "AI Assistant"},
						Level:        2,
					},
				},
			},
			{
				Name:        "discord_webhook",
				DisplayName: "Discord Webhook",
				Description: "Discord Webhook notification configuration - simple setup for single channel",
				Category:    CategoryProviders,
				Level:       0,
				Fields: []FieldMetadata{
					{
						Key:          "notifications.providers.discord_webhook.enabled",
						DisplayName:  "Enable Discord Webhook",
						Description:  "Whether to enable Discord Webhook notifications",
						Type:         TypeBool,
						Category:     CategoryProviders,
						Required:     false,
						DefaultValue: "false",
						Level:        1,
					},
					{
						Key:         "notifications.providers.discord_webhook.config.webhook_url",
						DisplayName: "Webhook URL",
						Description: "Discord channel Webhook URL for sending notifications",
						Type:        TypeSecret,
						Category:    CategoryProviders,
						Required:    false,
						Sensitive:   true,
						Examples:    []string{"https://discord.com/api/webhooks/123456789012345678/abcdefg..."},
						Validation:  "^https://discord\\.com/api/webhooks/\\d+/.+",
						Level:       2,
					},
					{
						Key:          "notifications.providers.discord_webhook.defaults.username",
						DisplayName:  "Webhook Username",
						Description:  "Username to display when sending messages via webhook",
						Type:         TypeString,
						Category:     CategoryProviders,
						Required:     false,
						DefaultValue: "Lai Bot",
						Examples:     []string{"Lai Bot", "Log Monitor", "AI Assistant"},
						Level:        2,
					},
				},
			},
			{
				Name:        "slack",
				DisplayName: "Slack Notifications",
				Description: "Slack Webhook notification configuration",
				Category:    CategoryProviders,
				Level:       0,
				Fields: []FieldMetadata{
					{
						Key:          "notifications.providers.slack.enabled",
						DisplayName:  "Enable Slack",
						Description:  "Whether to enable Slack notifications",
						Type:         TypeBool,
						Category:     CategoryProviders,
						Required:     false,
						DefaultValue: "false",
						Level:        1,
					},
					{
						Key:         "notifications.providers.slack.config.webhook_url",
						DisplayName: "Webhook URL",
						Description: "Slack channel Webhook URL",
						Type:        TypeSecret,
						Category:    CategoryProviders,
						Required:    false,
						Sensitive:   true,
						Examples:    []string{"https://hooks.slack.com/services/..."},
						Level:       2,
					},
				},
			},
			{
				Name:        "logging",
				DisplayName: "Logging Configuration",
				Description: "Application logging settings",
				Category:    CategoryLogging,
				Level:       0,
				Fields: []FieldMetadata{
					{
						Key:          "logging.level",
						DisplayName:  "Log Level",
						Description:  "Application logging level",
						Type:         TypeString,
						Category:     CategoryLogging,
						Required:     false,
						DefaultValue: "info",
						Examples:     []string{"debug", "info", "warn", "error"},
						Level:        1,
					},
				},
			},
		},
	}
}

// FindFieldMetadata finds metadata for a specific configuration key
func (cm *ConfigMetadata) FindFieldMetadata(key string) *FieldMetadata {
	for _, section := range cm.Sections {
		for _, field := range section.Fields {
			if field.Key == key {
				return &field
			}
		}
	}
	return nil
}

// GetSectionsByCategory returns all sections for a given category
func (cm *ConfigMetadata) GetSectionsByCategory(category Category) []ConfigSection {
	var sections []ConfigSection
	for _, section := range cm.Sections {
		if section.Category == category {
			sections = append(sections, section)
		}
	}
	return sections
}

// GetRequiredFields returns all required configuration fields
func (cm *ConfigMetadata) GetRequiredFields() []FieldMetadata {
	var required []FieldMetadata
	for _, section := range cm.Sections {
		for _, field := range section.Fields {
			if field.Required {
				required = append(required, field)
			}
		}
	}
	return required
}

// ValidateFieldValue validates a field value against its metadata
func (fm *FieldMetadata) ValidateFieldValue(value string) error {
	if fm.Required && value == "" {
		return fmt.Errorf("field %s is required", fm.Key)
	}

	if value == "" {
		return nil // Empty non-required fields are valid
	}

	switch fm.Type {
	case TypeInt:
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("field %s must be an integer", fm.Key)
		}
	case TypeBool:
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("field %s must be true or false", fm.Key)
		}
	case TypeDuration:
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("field %s must be a valid duration (e.g., 30s, 5m)", fm.Key)
		}
	case TypeString:
		// Check if there's a validation pattern for this field
		if fm.Validation != "" {
			if matched, err := regexp.MatchString(fm.Validation, value); err != nil || !matched {
				return fmt.Errorf("field %s has invalid format", fm.Key)
			}
		}
	case TypeSecret:
		// Check if there's a validation pattern for this field
		if fm.Validation != "" {
			if matched, err := regexp.MatchString(fm.Validation, value); err != nil || !matched {
				return fmt.Errorf("field %s has invalid format", fm.Key)
			}
		}
	case TypeStringList:
		// For string lists, we accept comma-separated values
		if strings.TrimSpace(value) != "" {
			parts := strings.Split(value, ",")
			for _, part := range parts {
				if strings.TrimSpace(part) == "" {
					return fmt.Errorf("field %s contains empty values in list", fm.Key)
				}
			}
		}
	}

	return nil
}

// GetDisplayValue returns a display-friendly version of the field value
func (fm *FieldMetadata) GetDisplayValue(value string) string {
	if fm.Sensitive && value != "" {
		// Show only first and last few characters for sensitive fields
		if len(value) <= 8 {
			return strings.Repeat("*", len(value))
		}
		return value[:3] + strings.Repeat("*", len(value)-6) + value[len(value)-3:]
	}
	return value
}
