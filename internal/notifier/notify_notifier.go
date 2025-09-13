package notifier

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/discord"
	"github.com/nikoksr/notify/service/slack"
	"github.com/nikoksr/notify/service/telegram"
	"github.com/shiquda/lai/internal/config"
	"github.com/shiquda/lai/internal/logger"
)

// NotifyNotifier is a universal notifier implemented using the notify library
// Supports all services provided by the notify library
type NotifyNotifier struct {
	notifyClient      *notify.Notify
	config            *config.NotificationsConfig
	enabledServices   map[string]bool
	serviceConfigs    map[string]config.ServiceConfig
}

// NewNotifyNotifier creates a new notify notifier
func NewNotifyNotifier(cfg *config.NotificationsConfig) (*NotifyNotifier, error) {
	nn := &NotifyNotifier{
		notifyClient:    notify.New(),
		config:          cfg,
		enabledServices: make(map[string]bool),
		serviceConfigs:  make(map[string]config.ServiceConfig),
	}

	// Setup all enabled notification services
	if err := nn.setupServices(); err != nil {
		return nil, fmt.Errorf("failed to setup notification services: %w", err)
	}

	return nn, nil
}

// setupServices sets up all enabled notification services
func (nn *NotifyNotifier) setupServices() error {
	if nn.config.Providers == nil {
		return fmt.Errorf("no notification providers configured")
	}

	for providerName, serviceConfig := range nn.config.Providers {
		if !serviceConfig.Enabled {
			logger.Infof("Provider %s is disabled, skipping", providerName)
			continue
		}

		// Save service configuration for later use
		nn.serviceConfigs[providerName] = serviceConfig

		if err := nn.setupProvider(providerName, serviceConfig); err != nil {
			logger.Warnf("Failed to setup %s service: %v", providerName, err)
			continue
		}

		nn.enabledServices[providerName] = true
		logger.Infof("Successfully enabled %s service", providerName)
	}

	// Setup fallback service
	if nn.config.Fallback != nil && nn.config.Fallback.Enabled {
		if err := nn.setupFallback(); err != nil {
			logger.Warnf("Failed to setup fallback service: %v", err)
		}
	}

	if len(nn.enabledServices) == 0 {
		return fmt.Errorf("no notification services could be enabled")
	}

	return nil
}

// setupProvider sets up a single notification service
func (nn *NotifyNotifier) setupProvider(providerName string, serviceConfig config.ServiceConfig) error {
	switch serviceConfig.Provider {
	case "telegram":
		return nn.setupTelegramService(serviceConfig)
	case "slack", "slack_webhook":
		return nn.setupSlackService(serviceConfig)
	case "discord", "discord_webhook":
		return nn.setupDiscordService(serviceConfig)
	case "email", "smtp", "gmail", "sendgrid", "mailgun":
		return nn.setupEmailService(serviceConfig)
	case "pushover":
		return nn.setupPushoverService(serviceConfig)
	case "twilio":
		return nn.setupTwilioService(serviceConfig)
	case "pagerduty":
		return nn.setupPagerDutyService(serviceConfig)
	case "dingtalk":
		return nn.setupDingTalkService(serviceConfig)
	case "wechat":
		return nn.setupWeChatService(serviceConfig)
	default:
		// Try to dynamically import other services
		return nn.setupGenericService(providerName, serviceConfig)
	}
}

// setupTelegramService sets up Telegram service using notify library
func (nn *NotifyNotifier) setupTelegramService(serviceConfig config.ServiceConfig) error {
	token, ok := serviceConfig.Config["bot_token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("telegram bot_token is required")
	}

	chatIDStr, ok := serviceConfig.Config["chat_id"].(string)
	if !ok || chatIDStr == "" {
		return fmt.Errorf("telegram chat_id is required")
	}

	// Create Telegram service
	telegramService, err := telegram.New(token)
	if err != nil {
		return fmt.Errorf("failed to create telegram service: %w", err)
	}

	// Parse chat IDs (supports multiple chat_ids)
	var chatIDs []string
	for _, chatID := range strings.Split(chatIDStr, ",") {
		chatID = strings.TrimSpace(chatID)
		if chatID != "" {
			chatIDs = append(chatIDs, chatID)
		}
	}

	// Add chat IDs as receivers
	for _, chatID := range chatIDs {
		// Convert string chat ID to int64
		chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid chat_id format '%s': %w", chatID, err)
		}
		telegramService.AddReceivers(chatIDInt)
	}

	// Add the service to notifyClient
	nn.notifyClient.UseServices(telegramService)

	return nil
}

// setupSlackService sets up Slack service
func (nn *NotifyNotifier) setupSlackService(serviceConfig config.ServiceConfig) error {
	if serviceConfig.Provider == "slack_webhook" {
		webhookURL, ok := serviceConfig.Config["webhook_url"].(string)
		if !ok || webhookURL == "" {
			return fmt.Errorf("slack webhook_url is required")
		}

		slackService := slack.New(webhookURL)
		nn.notifyClient.UseServices(slackService)
		return nil
	}

	// OAuth token method
	oauthToken, ok := serviceConfig.Config["oauth_token"].(string)
	if !ok || oauthToken == "" {
		return fmt.Errorf("slack oauth_token is required")
	}

	slackService := slack.New(oauthToken)

	// Add channels
	if channelIDs, ok := serviceConfig.Config["channel_ids"].([]interface{}); ok {
		for _, channelID := range channelIDs {
			if idStr, ok := channelID.(string); ok && idStr != "" {
				slackService.AddReceivers(idStr)
			}
		}
	}

	nn.notifyClient.UseServices(slackService)
	return nil
}

// setupDiscordService sets up Discord service
func (nn *NotifyNotifier) setupDiscordService(serviceConfig config.ServiceConfig) error {
	if serviceConfig.Provider == "discord_webhook" {
		webhookURL, ok := serviceConfig.Config["webhook_url"].(string)
		if !ok || webhookURL == "" {
			return fmt.Errorf("discord webhook_url is required")
		}

		_ = discord.New()
		// TODO: webhook support in notify library
		return fmt.Errorf("discord webhook support not yet implemented")
	}

	// Bot token method
	botToken, ok := serviceConfig.Config["bot_token"].(string)
	if !ok || botToken == "" {
		return fmt.Errorf("discord bot_token is required")
	}

	discordService := discord.New()
	if err := discordService.AuthenticateWithBotToken(botToken); err != nil {
		return fmt.Errorf("failed to authenticate with discord: %w", err)
	}

	// Add channels
	if channelIDs, ok := serviceConfig.Config["channel_ids"].([]interface{}); ok {
		for _, channelID := range channelIDs {
			if idStr, ok := channelID.(string); ok && idStr != "" {
				discordService.AddReceivers(idStr)
			}
		}
	}

	nn.notifyClient.UseServices(discordService)
	return nil
}

// setupEmailService sets up Email service
func (nn *NotifyNotifier) setupEmailService(serviceConfig config.ServiceConfig) error {
	switch serviceConfig.Provider {
	case "sendgrid":
		return nn.setupSendGridService(serviceConfig)
	case "mailgun":
		return nn.setupMailgunService(serviceConfig)
	case "smtp", "gmail":
		return nn.setupSMTPService(serviceConfig)
	default:
		return fmt.Errorf("unsupported email provider: %s", serviceConfig.Provider)
	}
}

// setupSendGridService sets up SendGrid email service
func (nn *NotifyNotifier) setupSendGridService(serviceConfig config.ServiceConfig) error {
	apiKey, ok := serviceConfig.Config["api_key"].(string)
	if !ok || apiKey == "" {
		return fmt.Errorf("sendgrid api_key is required")
	}

	fromEmail, ok := serviceConfig.Config["from_email"].(string)
	if !ok || fromEmail == "" {
		return fmt.Errorf("sendgrid from_email is required")
	}

	_ = "Lai Bot"
	if _, ok := serviceConfig.Config["from_name"].(string); ok {
		// from name configured but not used yet
	}

	// Note: sendgrid service
	// sendgridService := sendgrid.New(apiKey, fromEmail, fromName)
	// nn.notifyClient.UseServices(sendgridService)

	return fmt.Errorf("sendgrid service requires additional dependency: github.com/nikoksr/notify/service/sendgrid")
}

// setupMailgunService sets up Mailgun email service
func (nn *NotifyNotifier) setupMailgunService(serviceConfig config.ServiceConfig) error {
	apiKey, ok := serviceConfig.Config["api_key"].(string)
	if !ok || apiKey == "" {
		return fmt.Errorf("mailgun api_key is required")
	}

	domain, ok := serviceConfig.Config["domain"].(string)
	if !ok || domain == "" {
		return fmt.Errorf("mailgun domain is required")
	}

	// Note: mailgun service
	// mailgunService := mailgun.New(apiKey, domain)
	// nn.notifyClient.UseServices(mailgunService)

	return fmt.Errorf("mailgun service requires additional dependency: github.com/nikoksr/notify/service/mailgun")
}

// setupSMTPService sets up SMTP email service
func (nn *NotifyNotifier) setupSMTPService(serviceConfig config.ServiceConfig) error {
	// Validate required fields for SMTP configuration
	var host string
	if h, ok := serviceConfig.Config["host"].(string); ok && h != "" {
		host = h
	} else if h, ok := serviceConfig.Config["smtp_host"].(string); ok && h != "" {
		host = h
	} else {
		return fmt.Errorf("smtp host is required")
	}

	// Validate port (will be properly parsed in createEmailNotifier)
	var port int
	if p, ok := serviceConfig.Config["port"].(int); ok {
		port = p
	} else if p, ok := serviceConfig.Config["port"].(string); ok {
		if parsedPort, err := strconv.Atoi(p); err == nil {
			port = parsedPort
		} else {
			port = 587 // default
		}
	} else if p, ok := serviceConfig.Config["smtp_port"].(int); ok {
		port = p
	} else if p, ok := serviceConfig.Config["smtp_port"].(string); ok {
		if parsedPort, err := strconv.Atoi(p); err == nil {
			port = parsedPort
		} else {
			port = 587 // default
		}
	} else {
		port = 587 // default SMTP port
	}

	username, ok := serviceConfig.Config["username"].(string)
	if !ok || username == "" {
		return fmt.Errorf("smtp username is required")
	}

	password, ok := serviceConfig.Config["password"].(string)
	if !ok || password == "" {
		return fmt.Errorf("smtp password is required")
	}

	// Validate from email
	fromEmail, ok := serviceConfig.Config["from_email"].(string)
	if !ok || fromEmail == "" {
		fromEmail = username // default to username
	}

	// Validate recipient emails
	var toEmails []string
	if recipients, ok := serviceConfig.Config["to_emails"].([]interface{}); ok {
		for _, recipient := range recipients {
			if email, ok := recipient.(string); ok && email != "" {
				toEmails = append(toEmails, email)
			}
		}
	} else if recipient, ok := serviceConfig.Config["to_emails"].(string); ok && recipient != "" {
		toEmails = []string{recipient}
	}

	if len(toEmails) == 0 {
		return fmt.Errorf("smtp to_emails is required")
	}

	// Validate subject exists
	subject := "Log Summary Notification"
	if s, ok := serviceConfig.Config["subject"].(string); ok && s != "" {
		subject = s
	}

	// Validate TLS setting
	useTLS := true // default to TLS
	if tls, ok := serviceConfig.Config["use_tls"].(bool); ok {
		useTLS = tls
	}

	// Suppress unused variable warnings (these are validated above)
	_, _, _, _ = host, port, subject, useTLS

	// Note: We don't need to create the email notifier here since we handle email specially in TestProvider
	// The email notifier will be created on-demand when needed

	// Add to enabled services
	nn.enabledServices["email"] = true
	nn.serviceConfigs["email"] = serviceConfig

	return nil
}

// setupPushoverService sets up Pushover service
func (nn *NotifyNotifier) setupPushoverService(serviceConfig config.ServiceConfig) error {
	token, ok := serviceConfig.Config["token"].(string)
	if !ok || token == "" {
		return fmt.Errorf("pushover token is required")
	}

	user, ok := serviceConfig.Config["user"].(string)
	if !ok || user == "" {
		return fmt.Errorf("pushover user is required")
	}

	// Note: pushover service
	// pushoverService := pushover.New(token, user)
	// nn.notifyClient.UseServices(pushoverService)

	return fmt.Errorf("pushover service requires additional dependency: github.com/nikoksr/notify/service/pushover")
}

// setupTwilioService sets up Twilio SMS service
func (nn *NotifyNotifier) setupTwilioService(serviceConfig config.ServiceConfig) error {
	accountSid, ok := serviceConfig.Config["account_sid"].(string)
	if !ok || accountSid == "" {
		return fmt.Errorf("twilio account_sid is required")
	}

	authToken, ok := serviceConfig.Config["auth_token"].(string)
	if !ok || authToken == "" {
		return fmt.Errorf("twilio auth_token is required")
	}

	fromNumber, ok := serviceConfig.Config["from_number"].(string)
	if !ok || fromNumber == "" {
		return fmt.Errorf("twilio from_number is required")
	}

	// Note: twilio service
	// twilioService := twilio.New(accountSid, authToken, fromNumber)
	// nn.notifyClient.UseServices(twilioService)

	return fmt.Errorf("twilio service requires additional dependency: github.com/nikoksr/notify/service/twilio")
}

// setupPagerDutyService sets up PagerDuty service
func (nn *NotifyNotifier) setupPagerDutyService(serviceConfig config.ServiceConfig) error {
	routingKey, ok := serviceConfig.Config["routing_key"].(string)
	if !ok || routingKey == "" {
		return fmt.Errorf("pagerduty routing_key is required")
	}

	// Note: pagerduty service
	// pagerdutyService := pagerduty.New(routingKey)
	// nn.notifyClient.UseServices(pagerdutyService)

	return fmt.Errorf("pagerduty service requires additional dependency: github.com/nikoksr/notify/service/pagerduty")
}

// setupDingTalkService sets up DingTalk service
func (nn *NotifyNotifier) setupDingTalkService(serviceConfig config.ServiceConfig) error {
	accessToken, ok := serviceConfig.Config["access_token"].(string)
	if !ok || accessToken == "" {
		return fmt.Errorf("dingtalk access_token is required")
	}

	// Note: dingtalk service
	// dingtalkService := dingtalk.New(accessToken)
	// nn.notifyClient.UseServices(dingtalkService)

	return fmt.Errorf("dingtalk service requires additional dependency: github.com/nikoksr/notify/service/dingtalk")
}

// setupWeChatService sets up WeChat service
func (nn *NotifyNotifier) setupWeChatService(serviceConfig config.ServiceConfig) error {
	corpID, ok := serviceConfig.Config["corp_id"].(string)
	if !ok || corpID == "" {
		return fmt.Errorf("wechat corp_id is required")
	}

	corpSecret, ok := serviceConfig.Config["corp_secret"].(string)
	if !ok || corpSecret == "" {
		return fmt.Errorf("wechat corp_secret is required")
	}

	agentID, ok := serviceConfig.Config["agent_id"].(string)
	if !ok || agentID == "" {
		return fmt.Errorf("wechat agent_id is required")
	}

	// Note: wechat service
	// wechatService := wechat.New(corpID, corpSecret, agentID)
	// nn.notifyClient.UseServices(wechatService)

	return fmt.Errorf("wechat service requires additional dependency: github.com/nikoksr/notify/service/wechat")
}

// setupGenericService sets up generic service (dynamic import)
func (nn *NotifyNotifier) setupGenericService(providerName string, serviceConfig config.ServiceConfig) error {
	// Dynamic service import logic can be implemented here
	// Currently returns error to prompt user to install required dependencies
	return fmt.Errorf("provider '%s' is not supported. Please check if the service is available in the notify library", providerName)
}

// setupFallback sets up fallback service
func (nn *NotifyNotifier) setupFallback() error {
	if nn.config.Fallback == nil || !nn.config.Fallback.Enabled {
		return nil
	}

	fallbackConfig := config.ServiceConfig{
		Enabled:  true,
		Provider: nn.config.Fallback.Provider,
		Config:   nn.config.Fallback.Config,
		Defaults: make(map[string]interface{}),
	}

	return nn.setupProvider("fallback", fallbackConfig)
}

// SendLogSummary sends a log summary
func (nn *NotifyNotifier) SendLogSummary(ctx context.Context, filePath, summary string) error {
	if !nn.IsEnabled() {
		return fmt.Errorf("no notification channels enabled")
	}

	var errors []error

	// Special handling for email service since it's not using the notify library
	if nn.enabledServices["email"] {
		serviceConfig, exists := nn.serviceConfigs["email"]
		if !exists {
			errors = append(errors, fmt.Errorf("email service configuration not found"))
		} else {
			// Create email notifier for sending
			emailNotifier, err := nn.createEmailNotifier(serviceConfig)
			if err != nil {
				errors = append(errors, fmt.Errorf("failed to create email notifier: %w", err))
			} else {
				// Use the original EmailNotifier's SendLogSummary method for proper HTML rendering
				if err := emailNotifier.SendLogSummary(filePath, summary); err != nil {
					errors = append(errors, fmt.Errorf("failed to send email notification: %w", err))
				}
			}
		}
	}

	// Check if we have any services that use the notify library (including Telegram now)
	hasNotifyServices := false
	for service := range nn.enabledServices {
		if service != "email" { // Only skip email
			hasNotifyServices = true
			break
		}
	}

	if hasNotifyServices {
		// Format message for notify library services
		// For Telegram, we need to convert Markdown to a format it can handle
		var message string

		if nn.enabledServices["telegram"] {
			// Convert Markdown to HTML format for Telegram
			htmlFormattedSummary := nn.convertMarkdownToTelegramHTML(summary)
			htmlSafeFilePath := nn.makeHTMLSafe(filePath)

			message = fmt.Sprintf("🚨 <b>Log Summary Notification</b>\n\n📁 <b>File:</b> %s\n⏰ <b>Time:</b> %s\n\n📋 <b>Summary:</b>\n%s",
				htmlSafeFilePath, getCurrentTimeNotify(), htmlFormattedSummary)
		} else {
			// Use simple plain text format for other services
			plainSummary := nn.stripMarkdownFormatting(summary)
			message = fmt.Sprintf("Log Summary: %s", plainSummary)
		}

		// Send to all notify library services
		if err := nn.notifyClient.Send(ctx, "Log Summary Notification", message); err != nil {
			errors = append(errors, fmt.Errorf("failed to send notify library notification: %w", err))
		}
	}

	// Return combined errors if any
	if len(errors) > 0 {
		var errorMessages []string
		for _, err := range errors {
			errorMessages = append(errorMessages, err.Error())
		}
		return fmt.Errorf("notification errors: %s", strings.Join(errorMessages, "; "))
	}

	return nil
}

// SendMessage sends a plain message
func (nn *NotifyNotifier) SendMessage(ctx context.Context, message string) error {
	if !nn.IsEnabled() {
		return fmt.Errorf("no notification channels enabled")
	}

	// Special handling for email service since it's not using the notify library
	if nn.enabledServices["email"] {
		serviceConfig, exists := nn.serviceConfigs["email"]
		if !exists {
			return fmt.Errorf("email service configuration not found")
		}

		// Create email notifier for sending
		emailNotifier, err := nn.createEmailNotifier(serviceConfig)
		if err != nil {
			return fmt.Errorf("failed to create email notifier: %w", err)
		}

		// Convert the plain message to HTML email format
		// Replace newlines with <br> for better email formatting
		htmlMessage := strings.ReplaceAll(message, "\n", "<br>")
		emailMessage := fmt.Sprintf("<html><body><h2>📢 Lai Notification</h2><p>%s</p></body></html>", htmlMessage)
		return emailNotifier.SendMessage(emailMessage)
	}

	return nn.notifyClient.Send(ctx, "📢 Lai Notification", message)
}

// SendError sends an error message
func (nn *NotifyNotifier) SendError(ctx context.Context, filePath, errorMsg string) error {
	if !nn.IsEnabled() {
		return fmt.Errorf("no notification channels enabled")
	}

	message := nn.formatMessage("error", map[string]interface{}{
		"filePath": filePath,
		"errorMsg": errorMsg,
		"time":     getCurrentTimeNotify(),
	})

	// Special handling for email service since it's not using the notify library
	if nn.enabledServices["email"] {
		serviceConfig, exists := nn.serviceConfigs["email"]
		if !exists {
			return fmt.Errorf("email service configuration not found")
		}

		// Create email notifier for sending
		emailNotifier, err := nn.createEmailNotifier(serviceConfig)
		if err != nil {
			return fmt.Errorf("failed to create email notifier: %w", err)
		}

		// Convert the formatted message to HTML email format
		// Replace newlines with <br> for better email formatting
		htmlMessage := strings.ReplaceAll(message, "\n", "<br>")
		emailMessage := fmt.Sprintf("<html><body><h2>🚨 Critical Error Alert</h2><p>%s</p></body></html>", htmlMessage)
		return emailNotifier.SendMessage(emailMessage)
	}

	return nn.notifyClient.Send(ctx, "🚨 Critical Error Alert", message)
}

// TestProvider tests a specific provider
func (nn *NotifyNotifier) TestProvider(ctx context.Context, providerName string, message string) error {
	if !nn.enabledServices[providerName] {
		return fmt.Errorf("provider %s is not enabled", providerName)
	}

	// Special handling for email service since it's not using the notify library
	if providerName == "email" {
		serviceConfig, exists := nn.serviceConfigs[providerName]
		if !exists {
			return fmt.Errorf("email service configuration not found")
		}

		// Create a temporary email notifier for testing
		emailNotifier, err := nn.createEmailNotifier(serviceConfig)
		if err != nil {
			return fmt.Errorf("failed to create email notifier: %w", err)
		}

		testMessage := fmt.Sprintf("🧪 Test Message from Lai\n\nProvider: %s\nTime: %s\nMessage: %s",
			providerName, getCurrentTimeNotify(), message)

		return emailNotifier.SendMessage(testMessage)
	}

	testMessage := fmt.Sprintf("🧪 Test Message from Lai\n\nProvider: %s\nTime: %s\nMessage: %s",
		providerName, getCurrentTimeNotify(), message)

	return nn.notifyClient.Send(ctx, "🧪 Lai Test Notification", testMessage)
}

// createEmailNotifier creates an EmailNotifier from service configuration
func (nn *NotifyNotifier) createEmailNotifier(serviceConfig config.ServiceConfig) (*EmailNotifier, error) {
	// Get SMTP host
	var host string
	if h, ok := serviceConfig.Config["host"].(string); ok && h != "" {
		host = h
	} else if h, ok := serviceConfig.Config["smtp_host"].(string); ok && h != "" {
		host = h
	} else {
		return nil, fmt.Errorf("smtp host is required")
	}

	// Get SMTP port
	var port int
	if p, ok := serviceConfig.Config["port"].(int); ok {
		port = p
	} else if p, ok := serviceConfig.Config["port"].(string); ok {
		if parsedPort, err := strconv.Atoi(p); err == nil {
			port = parsedPort
		} else {
			port = 587
		}
	} else if p, ok := serviceConfig.Config["smtp_port"].(int); ok {
		port = p
	} else if p, ok := serviceConfig.Config["smtp_port"].(string); ok {
		if parsedPort, err := strconv.Atoi(p); err == nil {
			port = parsedPort
		} else {
			port = 587
		}
	} else {
		port = 587
	}

	// Get credentials
	username, ok := serviceConfig.Config["username"].(string)
	if !ok || username == "" {
		return nil, fmt.Errorf("smtp username is required")
	}

	password, ok := serviceConfig.Config["password"].(string)
	if !ok || password == "" {
		return nil, fmt.Errorf("smtp password is required")
	}

	// Get from email
	fromEmail, ok := serviceConfig.Config["from_email"].(string)
	if !ok || fromEmail == "" {
		fromEmail = username
	}

	// Get recipient emails
	var toEmails []string
	if recipients, ok := serviceConfig.Config["to_emails"].([]interface{}); ok {
		for _, recipient := range recipients {
			if email, ok := recipient.(string); ok && email != "" {
				toEmails = append(toEmails, email)
			}
		}
	} else if recipient, ok := serviceConfig.Config["to_emails"].(string); ok && recipient != "" {
		toEmails = []string{recipient}
	}

	if len(toEmails) == 0 {
		return nil, fmt.Errorf("smtp to_emails is required")
	}

	// Get subject
	subject := "Log Summary Notification"
	if s, ok := serviceConfig.Config["subject"].(string); ok && s != "" {
		subject = s
	}

	// Get TLS setting
	useTLS := true
	if tls, ok := serviceConfig.Config["use_tls"].(bool); ok {
		useTLS = tls
	}

	return NewEmailNotifier(host, port, username, password, fromEmail, toEmails, subject, useTLS, nil), nil
}

// IsServiceEnabled checks if a specific service is enabled
func (nn *NotifyNotifier) IsServiceEnabled(serviceName string) bool {
	return nn.enabledServices[serviceName]
}

// GetServiceConfig returns the configuration for a specific service
func (nn *NotifyNotifier) GetServiceConfig(serviceName string) (map[string]interface{}, bool) {
	serviceConfig, exists := nn.serviceConfigs[serviceName]
	if !exists {
		return nil, false
	}

	// Convert ServiceConfig to map[string]interface{}
	config := make(map[string]interface{})
	config["enabled"] = serviceConfig.Enabled
	config["provider"] = serviceConfig.Provider

	// Copy config values
	for k, v := range serviceConfig.Config {
		config[k] = v
	}

	// Copy defaults
	for k, v := range serviceConfig.Defaults {
		config[k] = v
	}

	return config, true
}

// IsEnabled checks if any notification channels are enabled
func (nn *NotifyNotifier) IsEnabled() bool {
	return len(nn.enabledServices) > 0
}

// GetEnabledChannels returns the list of enabled notification channels
func (nn *NotifyNotifier) GetEnabledChannels() []string {
	channels := make([]string, 0, len(nn.enabledServices))
	for channel := range nn.enabledServices {
		channels = append(channels, channel)
	}
	return channels
}

// formatMessage formats messages considering different service characteristics
func (nn *NotifyNotifier) formatMessage(msgType string, data map[string]interface{}) string {
	switch msgType {
	case "log_summary":
		return fmt.Sprintf("🚨 *Log Summary Notification*\n\n📁 *File:* %s\n⏰ *Time:* %s\n\n📋 *Summary:*\n%s",
			data["filePath"], data["time"], data["summary"])
	case "error":
		return fmt.Sprintf("🚨 *Critical Error Alert*\n\n📁 *File:* %s\n⏰ *Time:* %s\n\n💥 *Error Details:*\n%s",
			data["filePath"], data["time"], data["errorMsg"])
	default:
		return fmt.Sprintf("📢 *Notification*\n\n⏰ *Time:* %s\n\n📝 *Message:*\n%s",
			data["time"], data["message"])
	}
}

// convertMarkdownToTelegramHTML converts common Markdown formats to Telegram HTML
func (nn *NotifyNotifier) convertMarkdownToTelegramHTML(text string) string {
	// First escape HTML special characters to prevent conflicts
	text = nn.makeHTMLSafe(text)

	// Convert **bold** to <b>bold</b>
	text = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(text, "<b>$1</b>")

	// Convert single *italic* to <i>italic</i> (we do this after ** to avoid conflicts)
	text = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(text, "<i>$1</i>")

	// Convert _italic_ to <i>italic</i>
	text = regexp.MustCompile(`_([^_]+)_`).ReplaceAllString(text, "<i>$1</i>")

	// Convert `code` to <code>code</code>
	text = regexp.MustCompile("`([^`]+)`").ReplaceAllString(text, "<code>$1</code>")

	// Convert ```code block``` to <pre>code block</pre>
	text = regexp.MustCompile("```([^`]+)```").ReplaceAllString(text, "<pre>$1</pre>")

	// Convert [link text](url) to <a href="url">link text</a>
	text = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`).ReplaceAllString(text, `<a href="$2">$1</a>`)

	// Convert ### Header to <b>Header</b> (process line by line)
	lines := strings.Split(text, "\n")
	var result strings.Builder

	for _, line := range lines {
		// Handle markdown headers
		if matched := regexp.MustCompile(`^#{1,6}\s+(.+)$`).FindStringSubmatch(line); matched != nil {
			result.WriteString("<b>" + matched[1] + "</b>\n")
			continue
		}

		// Handle markdown list items (- item or * item)
		if matched := regexp.MustCompile(`^[\s]*[-*+]\s+(.+)$`).FindStringSubmatch(line); matched != nil {
			result.WriteString("• " + matched[1] + "\n")
			continue
		}

		// Handle numbered lists (1. item, 2. item, etc.)
		if matched := regexp.MustCompile(`^[\s]*\d+\.\s+(.+)$`).FindStringSubmatch(line); matched != nil {
			result.WriteString("• " + matched[1] + "\n")
			continue
		}

		// Handle horizontal rules (---, ***, ___)
		if regexp.MustCompile(`^[\s]*[-*_]{3,}[\s]*$`).MatchString(line) {
			result.WriteString("━━━━━━━━━━━━━━━━━━━━\n")
			continue
		}

		// Regular line
		result.WriteString(line + "\n")
	}

	finalText := result.String()
	// Remove trailing newline if exists
	if strings.HasSuffix(finalText, "\n") {
		finalText = finalText[:len(finalText)-1]
	}

	return finalText
}

// makeHTMLSafe escapes HTML special characters to prevent parsing errors
func (nn *NotifyNotifier) makeHTMLSafe(text string) string {
	// Escape HTML special characters
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&#39;")
	return text
}

// stripMarkdownFormatting removes all markdown formatting to ensure plain text
func (nn *NotifyNotifier) stripMarkdownFormatting(text string) string {
	// Remove markdown headers
	text = regexp.MustCompile(`^#{1,6}\s+`).ReplaceAllString(text, "")

	// Remove bold formatting **text** and *text*
	text = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(text, "$1")

	// Remove italic formatting _text_
	text = regexp.MustCompile(`_([^_]+)_`).ReplaceAllString(text, "$1")

	// Remove code formatting `code` and ```code```
	text = regexp.MustCompile("```([^`]+)```").ReplaceAllString(text, "$1")
	text = regexp.MustCompile("`([^`]+)`").ReplaceAllString(text, "$1")

	// Remove links [text](url) -> text
	text = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`).ReplaceAllString(text, "$1")

	// Remove any remaining special characters that might cause issues
	// Keep only safe characters, emojis, and basic punctuation
	text = regexp.MustCompile(`[\\*_\[\]()~|<>]`).ReplaceAllString(text, "")

	return text
}

// getCurrentTimeNotify gets current time (notify_notifier specific)
func getCurrentTimeNotify() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
