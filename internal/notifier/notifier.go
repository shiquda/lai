package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TelegramNotifier struct {
	botToken string
	chatID   string
	client   *http.Client
}

type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func NewTelegramNotifier(botToken, chatID string) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatID,
		client:   &http.Client{},
	}
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
		return fmt.Errorf("Telegram API request failed with status: %s", resp.Status)
	}

	return nil
}

func (t *TelegramNotifier) SendLogSummary(filePath, summary string) error {
	message := fmt.Sprintf(`🚨 *Log Summary Notification*

📁 File: %s
⏰ Time: %s

📋 Summary:
%s`, filePath, getCurrentTime(), summary)

	return t.SendMessage(message)
}

func getCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}