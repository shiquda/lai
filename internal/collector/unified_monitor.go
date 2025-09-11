package collector

import (
	"fmt"
	"strings"
	"time"

	"github.com/shiquda/lai/internal/config"
	"github.com/shiquda/lai/internal/logger"
	"github.com/shiquda/lai/internal/notifier"
	"github.com/shiquda/lai/internal/platform"
	"github.com/shiquda/lai/internal/summarizer"
)

// MonitorConfig represents unified monitoring configuration
type MonitorConfig struct {
	Source           MonitorSource
	LineThreshold    int
	CheckInterval    time.Duration
	ChatID           string
	Language         string
	ErrorOnlyMode    bool
	FinalSummary     bool
	FinalSummaryOnly bool
	OpenAI           config.OpenAIConfig
	Notifications    config.NotificationsConfig
}

// BuildMonitorConfig builds unified monitoring configuration
func BuildMonitorConfig(source MonitorSource, lineThreshold *int, checkInterval *time.Duration, chatID *string, workingDir string, finalSummary *bool, errorOnlyMode *bool, finalSummaryOnly *bool) (*MonitorConfig, error) {
	// Ensure global config exists
	if err := config.EnsureGlobalConfig(); err != nil {
		return nil, fmt.Errorf("failed to ensure global config: %w", err)
	}

	// Load global configuration
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// Build unified monitoring configuration
	cfg := &MonitorConfig{
		Source:           source,
		LineThreshold:    globalConfig.Defaults.LineThreshold,
		CheckInterval:    globalConfig.Defaults.CheckInterval,
		Language:         globalConfig.Defaults.Language,
		FinalSummary:     globalConfig.Defaults.FinalSummary,
		FinalSummaryOnly: globalConfig.Defaults.FinalSummaryOnly,
		ErrorOnlyMode:    globalConfig.Defaults.ErrorOnlyMode,
		OpenAI:           globalConfig.Notifications.OpenAI,
		Notifications:    globalConfig.Notifications,
	}

	// Apply command line parameter overrides
	if lineThreshold != nil {
		cfg.LineThreshold = *lineThreshold
	}
	if checkInterval != nil {
		cfg.CheckInterval = *checkInterval
	}
	if chatID != nil {
		cfg.ChatID = *chatID
	}
	if finalSummary != nil {
		cfg.FinalSummary = *finalSummary
	}
	if finalSummaryOnly != nil {
		cfg.FinalSummaryOnly = *finalSummaryOnly
	}
	if errorOnlyMode != nil {
		cfg.ErrorOnlyMode = *errorOnlyMode
	}

	// If no ChatID specified, use the default one from Telegram provider
	if cfg.ChatID == "" {
		if telegramProvider, exists := cfg.Notifications.Providers["telegram"]; exists {
			if chatID, ok := telegramProvider.Config["chat_id"].(string); ok {
				cfg.ChatID = chatID
			}
		}
	}

	return cfg, nil
}

// Validate validates the unified monitoring configuration
func (c *MonitorConfig) Validate() error {
	if c.Source == nil {
		return fmt.Errorf("monitor source is required")
	}

	if c.OpenAI.APIKey == "" {
		return fmt.Errorf("openai.api_key is required")
	}
	// Check if at least one notification provider is configured
	if len(c.Notifications.Providers) == 0 {
		return fmt.Errorf("at least one notification provider must be configured")
	}

	// Check if Telegram is properly configured if enabled
	if telegramProvider, exists := c.Notifications.Providers["telegram"]; exists && telegramProvider.Enabled {
		if token, ok := telegramProvider.Config["bot_token"].(string); !ok || token == "" {
			return fmt.Errorf("telegram.bot_token is required when telegram provider is enabled")
		}
		if chatID, ok := telegramProvider.Config["chat_id"].(string); !ok || chatID == "" {
			return fmt.Errorf("telegram.chat_id is required when telegram provider is enabled")
		}
	}
	if c.ChatID == "" {
		return fmt.Errorf("chat_id is required (set via --chat-id or defaults.chat_id in global config)")
	}
	return nil
}

// UnifiedMonitor represents a unified monitoring system
type UnifiedMonitor struct {
	config     *MonitorConfig
	collector  LogCollector
	summarizer *summarizer.OpenAIClient
	notifiers  []notifier.Notifier
}

// NewUnifiedMonitor creates a new unified monitor
func NewUnifiedMonitor(cfg *MonitorConfig) (*UnifiedMonitor, error) {
	// Create OpenAI client
	openaiClient := summarizer.NewOpenAIClient(cfg.OpenAI.APIKey, cfg.OpenAI.BaseURL, cfg.OpenAI.Model)

	// Create notifiers
	// Create a temporary config object for notifier creation
	tempConfig := &config.Config{
		Notifications: cfg.Notifications,
	}

	notifiers, err := notifier.CreateNotifiers(tempConfig, []string{})
	if err != nil {
		return nil, fmt.Errorf("failed to create notifiers: %w", err)
	}

	// Create appropriate collector based on source type
	var collector LogCollector
	identifier := cfg.Source.GetIdentifier()

	// Check if this is a command source by prefix
	if strings.HasPrefix(identifier, "COMMAND_SOURCE:") {
		// Extract the original command string
		originalCmdStr := strings.TrimPrefix(identifier, "COMMAND_SOURCE:")

		// Parse the command string to extract command and args
		command, args, err := parseCommandString(originalCmdStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse command: %w", err)
		}

		collector = NewStreamCollector(command, args, cfg.LineThreshold, cfg.CheckInterval, cfg.FinalSummary)
	} else {
		// Regular file monitoring
		collector = New(identifier, cfg.LineThreshold, cfg.CheckInterval)
	}

	return &UnifiedMonitor{
		config:     cfg,
		collector:  collector,
		summarizer: openaiClient,
		notifiers:  notifiers,
	}, nil
}

// Start begins monitoring
func (m *UnifiedMonitor) Start() error {
	// Set trigger handler
	m.collector.SetTriggerHandler(func(newContent string) error {
		logger.Printf("Changes detected, processing...\n")

		if m.config.ErrorOnlyMode {
			// Error-only mode: first check if content contains errors
			analysis, err := m.summarizer.AnalyzeForErrors(newContent, m.config.Language)
			if err != nil {
				return fmt.Errorf("failed to analyze errors: %w", err)
			}

			if !analysis.HasError {
				logger.Printf("No errors detected, skipping notification (error-only mode)\n")
				return nil
			}

			logger.Printf("Error detected (severity: %s), sending notification\n", analysis.Severity)
			if err := m.sendToAllNotifiers(analysis.Summary); err != nil {
				return fmt.Errorf("failed to send notification: %w", err)
			}
		} else {
			// Normal mode: generate summary and send notification
			logger.Printf("Generating summary...\n")
			summary, err := m.summarizer.Summarize(newContent, m.config.Language)
			if err != nil {
				return fmt.Errorf("failed to generate summary: %w", err)
			}

			if err := m.sendToAllNotifiers(summary); err != nil {
				return fmt.Errorf("failed to send notification: %w", err)
			}
		}

		// Note: Individual notification status is now logged in sendToAllNotifiers method
		return nil
	})

	// Display startup information
	logger.Printf("Starting monitoring: %s\n", m.config.Source.GetIdentifier())
	logger.Printf("Type: %s\n", m.config.Source.GetType())
	logger.Printf("Line threshold: %d lines\n", m.config.LineThreshold)
	logger.Printf("Check interval: %v\n", m.config.CheckInterval)
	if m.config.ErrorOnlyMode {
		logger.Printf("Error-only mode: ENABLED (will only notify on errors/exceptions)\n")
	} else {
		logger.Printf("Error-only mode: DISABLED (will notify on all changes)\n")
	}

	// Setup signal handling
	p := platform.New()
	sigChan := p.Signal.SetupShutdownSignals()

	// Run collector in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- m.collector.Start()
	}()

	// Wait for signal or error
	select {
	case <-sigChan:
		logger.Println("\nReceived stop signal, shutting down...")
		return nil
	case err := <-errChan:
		return err
	}
}

// Stop stops the monitoring
func (m *UnifiedMonitor) Stop() {
	if streamCollector, ok := m.collector.(*StreamCollector); ok {
		streamCollector.Stop()
	}
}

// sendToAllNotifiers sends a summary to all configured notifiers
func (m *UnifiedMonitor) sendToAllNotifiers(summary string) error {
	var errors []error
	var successfulNotifiers []string

	for _, n := range m.notifiers {
		if err := n.SendLogSummary(m.config.Source.GetIdentifier(), summary); err != nil {
			errors = append(errors, err)
			logger.Errorf("Failed to send notification to %s notifier: %v\n", n.Name(), err)
		} else {
			successfulNotifiers = append(successfulNotifiers, n.Name())
			logger.Printf("Notification sent successfully to %s notifier\n", n.Name())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("%d notifier(s) failed: %v", len(errors), errors)
	}

	return nil
}

// parseCommandString parses a command string into command and arguments
// This is a simplified version that handles quoted arguments
func parseCommandString(cmdStr string) (command string, args []string, err error) {
	var words []string
	var currentWord strings.Builder
	var inSingleQuote bool
	var inDoubleQuote bool
	var escapeNext bool

	for _, r := range cmdStr {
		if escapeNext {
			currentWord.WriteRune(r)
			escapeNext = false
			continue
		}

		switch r {
		case '\\':
			escapeNext = true
		case '\'':
			if inDoubleQuote {
				currentWord.WriteRune(r)
			} else {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if inSingleQuote {
				currentWord.WriteRune(r)
			} else {
				inDoubleQuote = !inDoubleQuote
			}
		case ' ', '\t':
			if inSingleQuote || inDoubleQuote {
				currentWord.WriteRune(r)
			} else {
				if currentWord.Len() > 0 {
					words = append(words, currentWord.String())
					currentWord.Reset()
				}
			}
		default:
			currentWord.WriteRune(r)
		}
	}

	// Add the last word
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	if escapeNext {
		return "", nil, fmt.Errorf("unterminated escape sequence")
	}
	if inSingleQuote {
		return "", nil, fmt.Errorf("unterminated single quote")
	}
	if inDoubleQuote {
		return "", nil, fmt.Errorf("unterminated double quote")
	}

	if len(words) == 0 {
		return "", nil, fmt.Errorf("no command found")
	}

	return words[0], words[1:], nil
}
