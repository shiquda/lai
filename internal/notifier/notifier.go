package notifier

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/shiquda/lai/internal/config"
)

// Notifier defines the interface for all notification implementations.
// Any notifier must implement these two methods to be compatible with the system.
type Notifier interface {
	// SendMessage sends a plain message to the notification channel
	SendMessage(message string) error
	
	// SendLogSummary sends a formatted log summary to the notification channel
	// This method typically uses templates to format the message consistently
	SendLogSummary(filePath, summary string) error
}

// TemplateData represents the data available in message templates.
// These fields can be used in custom notification templates.
type TemplateData struct {
	FilePath    string // Path to the log file being monitored
	Time        string // Timestamp when the notification was sent
	Summary     string // AI-generated log summary content
	ProcessName string // Name of the monitoring process (if set)
	LineCount   int    // Number of lines that triggered the notification
}

// getCurrentTime returns the current time in a standardized format.
// This format is used consistently across all notifications.
func getCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// RenderTemplate renders a message template with the given data.
// It provides fallback functionality - if custom template fails, it uses default templates.
//
// Parameters:
//   - templateName: Name of the template to render (e.g., "log_summary")
//   - data: TemplateData containing the dynamic values
//   - messageTemplates: Custom templates from configuration
//   - getDefaultTemplates: Function to get default templates as fallback
//
// Returns:
//   - Rendered template string
//   - Error if both custom and default templates fail
func RenderTemplate(templateName string, data TemplateData, messageTemplates map[string]string, getDefaultTemplates func() map[string]string) (string, error) {
	templateStr, exists := messageTemplates[templateName]
	if !exists {
		// Fall back to default template
		defaultTemplates := getDefaultTemplates()
		templateStr = defaultTemplates[templateName]
	}

	tmpl, err := template.New(templateName).Parse(templateStr)
	if err != nil {
		// Fall back to default template on parse error
		defaultTemplates := getDefaultTemplates()
		templateStr = defaultTemplates[templateName]
		tmpl, err = template.New(templateName).Parse(templateStr)
		if err != nil {
			return "", fmt.Errorf("failed to parse default template: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// SendLogSummary sends a log summary using the provided notifier and templates.
// This is a convenience function that handles template rendering and message sending.
//
// Parameters:
//   - notifier: The notifier implementation to use (Telegram, Email, etc.)
//   - filePath: Path to the log file
//   - summary: AI-generated log summary content
//   - messageTemplates: Custom templates from configuration
//   - getDefaultTemplates: Function to get default templates as fallback
//
// Returns:
//   - Error if template rendering or message sending fails
func SendLogSummary(notifier Notifier, filePath, summary string, messageTemplates map[string]string, getDefaultTemplates func() map[string]string) error {
	data := TemplateData{
		FilePath: filePath,
		Time:     getCurrentTime(),
		Summary:  summary,
	}

	message, err := RenderTemplate("log_summary", data, messageTemplates, getDefaultTemplates)
	if err != nil {
		return fmt.Errorf("failed to render log summary template: %w", err)
	}

	return notifier.SendMessage(message)
}

// CreateNotifiers creates notifiers based on the provided configuration.
// This is a factory function that creates notifier instances based on configuration
// and command-line specifications.
//
// Priority order for determining which notifiers to enable:
// 1. Command line specifications (--notifiers telegram,email)
// 2. Configuration file defaults (defaults.enabled_notifiers)
// 3. All configured notifiers (if no specification exists)
//
// Parameters:
//   - cfg: Configuration containing notifier settings
//   - enabledNotifiers: List of notifiers to enable (from command line)
//
// Returns:
//   - Slice of configured notifier instances
//   - Error if no valid notifiers can be created
func CreateNotifiers(cfg *config.Config, enabledNotifiers []string) ([]Notifier, error) {
	var notifiers []Notifier

	// Determine which notifiers to enable based on priority:
	// 1. Command line parameters have highest priority
	// 2. Configuration defaults have second priority
	// 3. If neither specified, enable all configured notifiers
	var notifiersToCheck []string
	
	if len(enabledNotifiers) > 0 {
		// Command line specification takes precedence
		notifiersToCheck = enabledNotifiers
	} else if len(cfg.EnabledNotifiers) > 0 {
		// Use configuration defaults
		notifiersToCheck = cfg.EnabledNotifiers
	} else {
		// Enable all available notifiers by default
		notifiersToCheck = []string{"telegram", "email"}
	}

	// Check each notifier for valid configuration
	enableTelegram := false
	enableEmail := false
	
	for _, notifierType := range notifiersToCheck {
		switch notifierType {
		case "telegram":
			if cfg.Telegram.BotToken != "" && cfg.Telegram.ChatID != "" {
				enableTelegram = true
			}
		case "email":
			if cfg.Email.SMTPHost != "" && cfg.Email.Username != "" && len(cfg.Email.ToEmails) > 0 {
				enableEmail = true
			}
		}
	}

	// Create Telegram notifier if enabled and configured
	if enableTelegram {
		telegramNotifier := NewTelegramNotifier(
			cfg.Telegram.BotToken,
			cfg.Telegram.ChatID,
			convertTelegramTemplates(cfg.Telegram.MessageTemplates),
		)
		notifiers = append(notifiers, telegramNotifier)
	}

	// Create Email notifier if enabled and configured
	if enableEmail {
		emailNotifier := NewEmailNotifier(
			cfg.Email.SMTPHost,
			cfg.Email.SMTPPort,
			cfg.Email.Username,
			cfg.Email.Password,
			cfg.Email.FromEmail,
			cfg.Email.ToEmails,
			cfg.Email.Subject,
			cfg.Email.UseTLS,
			convertEmailTemplates(cfg.Email.MessageTemplates),
		)
		notifiers = append(notifiers, emailNotifier)
	}

	if len(notifiers) == 0 {
		return nil, fmt.Errorf("no valid notifier configuration found")
	}

	return notifiers, nil
}

// convertTelegramTemplates converts TelegramMessageTemplates to map[string]string
func convertTelegramTemplates(templates config.TelegramMessageTemplates) map[string]string {
	if templates.LogSummary == "" {
		return nil
	}
	return map[string]string{
		"log_summary": templates.LogSummary,
	}
}

// convertEmailTemplates converts EmailMessageTemplates to map[string]string
func convertEmailTemplates(templates config.EmailMessageTemplates) map[string]string {
	if templates.LogSummary == "" {
		return nil
	}
	return map[string]string{
		"log_summary": templates.LogSummary,
	}
}

// UnifiedConfig represents the interface for unified configuration
type UnifiedConfig interface {
	GetEnabledNotifiers() []string
	GetTelegramConfig() config.TelegramConfig
	GetEmailConfig() config.EmailConfig
}

// CreateNotifiersForUnified creates notifiers based on the unified configuration.
// This is similar to CreateNotifiers but works with the new MonitorConfig.
func CreateNotifiersForUnified(cfg UnifiedConfig, enabledNotifiers []string) ([]Notifier, error) {
	var notifiers []Notifier

	// Determine which notifiers to enable based on priority:
	// 1. Command line parameters have highest priority
	// 2. Configuration defaults have second priority
	// 3. If neither specified, enable all configured notifiers
	var notifiersToCheck []string
	
	if len(enabledNotifiers) > 0 {
		// Command line specification takes precedence
		notifiersToCheck = enabledNotifiers
	} else if len(cfg.GetEnabledNotifiers()) > 0 {
		// Use configuration defaults
		notifiersToCheck = cfg.GetEnabledNotifiers()
	} else {
		// Enable all available notifiers by default
		notifiersToCheck = []string{"telegram", "email"}
	}

	// Check each notifier for valid configuration
	enableTelegram := false
	enableEmail := false
	
	for _, notifierType := range notifiersToCheck {
		switch notifierType {
		case "telegram":
			telegramConfig := cfg.GetTelegramConfig()
			if telegramConfig.BotToken != "" && telegramConfig.ChatID != "" {
				enableTelegram = true
			}
		case "email":
			emailConfig := cfg.GetEmailConfig()
			if emailConfig.SMTPHost != "" && emailConfig.Username != "" && len(emailConfig.ToEmails) > 0 {
				enableEmail = true
			}
		}
	}

	// Create Telegram notifier if enabled and configured
	if enableTelegram {
		telegramConfig := cfg.GetTelegramConfig()
		telegramNotifier := NewTelegramNotifier(
			telegramConfig.BotToken,
			telegramConfig.ChatID,
			convertTelegramTemplatesForUnified(telegramConfig),
		)
		notifiers = append(notifiers, telegramNotifier)
	}

	// Create Email notifier if enabled and configured
	if enableEmail {
		emailConfig := cfg.GetEmailConfig()
		emailNotifier := NewEmailNotifier(
			emailConfig.SMTPHost,
			emailConfig.SMTPPort,
			emailConfig.Username,
			emailConfig.Password,
			emailConfig.FromEmail,
			emailConfig.ToEmails,
			emailConfig.Subject,
			emailConfig.UseTLS,
			convertEmailTemplatesForUnified(emailConfig),
		)
		notifiers = append(notifiers, emailNotifier)
	}

	if len(notifiers) == 0 {
		return nil, fmt.Errorf("no valid notifier configuration found")
	}

	return notifiers, nil
}

// convertTelegramTemplatesForUnified converts TelegramConfig to map[string]string
func convertTelegramTemplatesForUnified(telegramConfig config.TelegramConfig) map[string]string {
	if telegramConfig.MessageTemplates.LogSummary == "" {
		return nil
	}
	return map[string]string{
		"log_summary": telegramConfig.MessageTemplates.LogSummary,
	}
}

// convertEmailTemplatesForUnified converts EmailConfig to map[string]string
func convertEmailTemplatesForUnified(emailConfig config.EmailConfig) map[string]string {
	if emailConfig.MessageTemplates.LogSummary == "" {
		return nil
	}
	return map[string]string{
		"log_summary": emailConfig.MessageTemplates.LogSummary,
	}
}
