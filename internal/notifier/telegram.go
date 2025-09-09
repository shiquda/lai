package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// TelegramNotifier implements the Notifier interface for Telegram notifications.
// It sends messages via the Telegram Bot API.
type TelegramNotifier struct {
	botToken         string                // Telegram bot token from @BotFather
	chatID           string                // Telegram chat ID or channel ID
	client           *http.Client          // HTTP client for API requests
	messageTemplates map[string]string     // Custom message templates
}

// TelegramMessage represents the JSON structure for Telegram Bot API messages.
type TelegramMessage struct {
	ChatID    string `json:"chat_id"`    // Target chat ID
	Text      string `json:"text"`      // Message text content
	ParseMode string `json:"parse_mode,omitempty"` // Message parsing mode (Markdown, HTML)
}

// NewTelegramNotifier creates a new TelegramNotifier instance.
//
// Parameters:
//   - botToken: Telegram bot token from @BotFather
//   - chatID: Target chat ID or channel ID
//   - templates: Custom message templates (nil to use defaults)
//
// Returns:
//   - Configured TelegramNotifier instance
func NewTelegramNotifier(botToken, chatID string, templates map[string]string) *TelegramNotifier {
	if templates == nil {
		templates = getTelegramDefaultTemplates()
	}
	return &TelegramNotifier{
		botToken:         botToken,
		chatID:           chatID,
		client:           &http.Client{},
		messageTemplates: templates,
	}
}

func (t *TelegramNotifier) SetClient(client *http.Client) {
	t.client = client
}

func (t *TelegramNotifier) SendMessage(message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	telegramMsg := TelegramMessage{
		ChatID:    t.chatID,
		Text:      message,
		ParseMode: "Markdown",
	}

	jsonData, err := json.Marshal(telegramMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api request failed with status: %s", resp.Status)
	}

	return nil
}

// getTelegramDefaultTemplates returns the default message templates
func getTelegramDefaultTemplates() map[string]string {
	return map[string]string{
		"log_summary": `🚨 *Log Summary Notification*

📁 File: {{.FilePath}}
⏰ Time: {{.Time}}

📋 Summary:
{{.Summary}}`,
	}
}

func (t *TelegramNotifier) SendLogSummary(filePath, summary string) error {
	return SendLogSummary(t, filePath, summary, t.messageTemplates, getTelegramDefaultTemplates)
}
