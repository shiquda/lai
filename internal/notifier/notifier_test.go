package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shiquda/lai/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type NotifierTestSuite struct {
	suite.Suite
	server   *httptest.Server
	notifier *TelegramNotifier
}

func (s *NotifierTestSuite) SetupTest() {
	s.server = httptest.NewServer(http.HandlerFunc(s.mockTelegramHandler))

	botToken := "test-bot-token"
	chatID := "-100123456789"

	s.notifier = NewTelegramNotifier(botToken, chatID, nil)
	s.notifier.client = s.server.Client()
}

func (s *NotifierTestSuite) TearDownTest() {
	if s.server != nil {
		s.server.Close()
	}
}

func (s *NotifierTestSuite) mockTelegramHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var telegramMsg TelegramMessage
	if err := json.NewDecoder(r.Body).Decode(&telegramMsg); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if telegramMsg.ChatID == "" || telegramMsg.Text == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if telegramMsg.ChatID == "error-chat-id" {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"ok": true,
		"result": map[string]interface{}{
			"message_id": 123,
			"date":       time.Now().Unix(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func TestNotifierSuite(t *testing.T) {
	suite.Run(t, new(NotifierTestSuite))
}

func TestNewTelegramNotifier(t *testing.T) {
	botToken := "123456:ABC-DEF1234ghIKl-zyx57W2v1u123ew11"
	chatID := "-100123456789"

	notifier := NewTelegramNotifier(botToken, chatID, nil)

	assert.NotNil(t, notifier)
	assert.Equal(t, botToken, notifier.botToken)
	assert.Equal(t, chatID, notifier.chatID)
	assert.NotNil(t, notifier.client)
}

func (s *NotifierTestSuite) TestSendMessage_Success() {
	err := s.sendMessageWithCustomURL("Hello, World!")

	assert.NoError(s.T(), err)
}

func (s *NotifierTestSuite) TestSendMessage_ServerError() {
	s.notifier.chatID = "error-chat-id"

	err := s.sendMessageWithCustomURL("Hello, World!")

	assert.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "Telegram API request failed")
}

func (s *NotifierTestSuite) TestSendMessage_InvalidJSON() {
	originalChatID := s.notifier.chatID
	defer func() { s.notifier.chatID = originalChatID }()

	s.notifier.chatID = ""

	err := s.sendMessageWithCustomURL("Hello, World!")

	assert.Error(s.T(), err)
}

func (s *NotifierTestSuite) TestSendLogSummary() {
	filePath := testutils.GetTestLogPath("test.log")
	summary := "Test log summary with important information"

	err := s.sendLogSummaryWithCustomURL(filePath, summary)

	assert.NoError(s.T(), err)
}

func (s *NotifierTestSuite) TestTelegramMessage_Structure() {
	msg := TelegramMessage{
		ChatID:    "-100123456789",
		Text:      "Test message",
		ParseMode: "Markdown",
	}

	jsonData, err := json.Marshal(msg)
	assert.NoError(s.T(), err)

	var unmarshaled TelegramMessage
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), msg.ChatID, unmarshaled.ChatID)
	assert.Equal(s.T(), msg.Text, unmarshaled.Text)
	assert.Equal(s.T(), msg.ParseMode, unmarshaled.ParseMode)
}

func (s *NotifierTestSuite) TestSendLogSummary_MessageFormat() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var telegramMsg TelegramMessage
		json.NewDecoder(r.Body).Decode(&telegramMsg)

		assert.Contains(s.T(), telegramMsg.Text, "🚨 *Log Summary Notification*")
		assert.Contains(s.T(), telegramMsg.Text, "📁 File:")
		assert.Contains(s.T(), telegramMsg.Text, "⏰ Time:")
		assert.Contains(s.T(), telegramMsg.Text, "📋 Summary:")
		assert.Contains(s.T(), telegramMsg.Text, testutils.GetTestLogPath("test.log"))
		assert.Contains(s.T(), telegramMsg.Text, "Test summary")
		assert.Equal(s.T(), "Markdown", telegramMsg.ParseMode)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer server.Close()

	notifier := NewTelegramNotifier("test-token", "-100123456789", nil)
	notifier.client = server.Client()

	err := s.sendLogSummaryWithCustomURLAndNotifier(notifier, server.URL, testutils.GetTestLogPath("test.log"), "Test summary")
	assert.NoError(s.T(), err)
}

func TestGetCurrentTime(t *testing.T) {
	timeStr := getCurrentTime()

	assert.NotEmpty(t, timeStr)

	parsedTime, err := time.Parse("2006-01-02 15:04:05", timeStr)
	assert.NoError(t, err)

	timeDiff := time.Since(parsedTime)
	assert.True(t, timeDiff < time.Minute, "Time should be recent")
}

func (s *NotifierTestSuite) sendMessageWithCustomURL(message string) error {
	url := fmt.Sprintf("%s/bot%s/sendMessage", s.server.URL, s.notifier.botToken)

	telegramMsg := TelegramMessage{
		ChatID:    s.notifier.chatID,
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

	resp, err := s.notifier.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram API request failed with status: %s", resp.Status)
	}

	return nil
}

func (s *NotifierTestSuite) sendLogSummaryWithCustomURL(filePath, summary string) error {
	message := fmt.Sprintf(`🚨 *Log Summary Notification*

📁 File: %s
⏰ Time: %s

📋 Summary:
%s`, filePath, getCurrentTime(), summary)

	return s.sendMessageWithCustomURL(message)
}

func (s *NotifierTestSuite) sendLogSummaryWithCustomURLAndNotifier(notifier *TelegramNotifier, baseURL, filePath, summary string) error {
	message := fmt.Sprintf(`🚨 *Log Summary Notification*

📁 File: %s
⏰ Time: %s

📋 Summary:
%s`, filePath, getCurrentTime(), summary)

	url := fmt.Sprintf("%s/bot%s/sendMessage", baseURL, notifier.botToken)

	telegramMsg := TelegramMessage{
		ChatID:    notifier.chatID,
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

	resp, err := notifier.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram API request failed with status: %s", resp.Status)
	}

	return nil
}
