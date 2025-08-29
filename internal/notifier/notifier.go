package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"
)

type TelegramNotifier struct {
	botToken         string
	chatID           string
	client           *http.Client
	messageTemplates map[string]string
}

// TemplateData represents the data available in message templates
type TemplateData struct {
	FilePath    string
	Time        string
	Summary     string
	ProcessName string
	LineCount   int
}

type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func NewTelegramNotifier(botToken, chatID string, templates map[string]string) *TelegramNotifier {
	if templates == nil {
		templates = getDefaultTemplates()
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

func (t *TelegramNotifier) SetBaseURL(baseURL string) {
	t.botToken = strings.Replace(t.botToken, "bot", "", 1) // Remove 'bot' prefix if present
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

// getDefaultTemplates returns the default message templates
func getDefaultTemplates() map[string]string {
	return map[string]string{
		"log_summary": `🚨 *Log Summary Notification*

📁 File: {{.FilePath}}
⏰ Time: {{.Time}}

📋 Summary:
{{.Summary}}`,
	}
}

// renderTemplate renders a message template with the given data
func (t *TelegramNotifier) renderTemplate(templateName string, data TemplateData) (string, error) {
	templateStr, exists := t.messageTemplates[templateName]
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

func (t *TelegramNotifier) SendLogSummary(filePath, summary string) error {
	data := TemplateData{
		FilePath: filePath,
		Time:     getCurrentTime(),
		Summary:  summary,
	}

	message, err := t.renderTemplate("log_summary", data)
	if err != nil {
		return fmt.Errorf("failed to render log summary template: %w", err)
	}

	return t.SendMessage(message)
}

func getCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
