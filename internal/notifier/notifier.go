package notifier

import (
	"bytes"
	"context"
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

// UnifiedNotifier defines the interface for the new unified notification system
// using the notify library. This provides context-aware operations and better error handling.
type UnifiedNotifier interface {
	// SendLogSummary sends a log summary with context
	SendLogSummary(ctx context.Context, filePath, summary string) error

	// SendMessage sends a plain message with context
	SendMessage(ctx context.Context, message string) error

	// SendError sends an error message with context
	SendError(ctx context.Context, filePath, errorMsg string) error

	// TestProvider tests a specific notification provider
	TestProvider(ctx context.Context, providerName string, message string) error

	// IsEnabled checks if any notification channels are enabled
	IsEnabled() bool

	// IsServiceEnabled checks if a specific service is enabled
	IsServiceEnabled(serviceName string) bool

	// GetEnabledChannels returns the list of enabled notification channels
	GetEnabledChannels() []string
}

// CreateUnifiedNotifier creates a new unified notifier using the notify library.
// This is the recommended factory function for new implementations.
// It now uses the new provider configuration system.
func CreateUnifiedNotifier(cfg *config.Config) (UnifiedNotifier, error) {
	// Check if we need to migrate legacy configuration
	if len(cfg.Notifications.Providers) == 0 {
		// Migrate legacy configuration to new format
		migratedConfig := config.MigrateToNewProviderConfig(&config.GlobalConfig{
			Notifications: cfg.Notifications,
		})
		cfg.Notifications = migratedConfig.Notifications
	}

	// Create the new notify notifier
	notifyNotifier, err := NewNotifyNotifier(&cfg.Notifications)
	if err != nil {
		return nil, fmt.Errorf("failed to create notify notifier: %w", err)
	}

	return notifyNotifier, nil
}

// CreateUnifiedNotifierForMonitor creates a unified notifier for monitor configuration
func CreateUnifiedNotifierForMonitor(monitorConfig interface{}) (UnifiedNotifier, error) {
	// Try to convert to different config types
	switch cfg := monitorConfig.(type) {
	case *config.Config:
		return CreateUnifiedNotifier(cfg)
	case *config.NotificationsConfig:
		return NewNotifyNotifier(cfg)
	default:
		return nil, fmt.Errorf("unsupported config type: %T", monitorConfig)
	}
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

// CreateNotifiers creates notifiers based on the provided configuration.
// This function now uses the new unified notification system.
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
	// For backward compatibility, create a unified notifier
	// This will automatically migrate legacy configuration if needed
	unifiedNotifier, err := CreateUnifiedNotifier(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create unified notifier: %w", err)
	}

	// Check if the unified notifier is enabled
	if !unifiedNotifier.IsEnabled() {
		return nil, fmt.Errorf("no notification channels enabled")
	}

	// Return a slice containing just the unified notifier
	// This maintains compatibility with the existing interface
	return []Notifier{&UniversalNotifier{unified: unifiedNotifier}}, nil
}

// UniversalNotifier is a wrapper that implements the legacy Notifier interface
// but uses the new UnifiedNotifier internally for backward compatibility
type UniversalNotifier struct {
	unified UnifiedNotifier
}

// SendMessage sends a plain message (compatible with legacy interface)
func (un *UniversalNotifier) SendMessage(message string) error {
	ctx := context.Background()
	return un.unified.SendMessage(ctx, message)
}

// SendLogSummary sends a log summary (compatible with legacy interface)
func (un *UniversalNotifier) SendLogSummary(filePath, summary string) error {
	ctx := context.Background()
	return un.unified.SendLogSummary(ctx, filePath, summary)
}
