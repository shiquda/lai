package notifier

import (
	"testing"

	"github.com/shiquda/lai/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestTemplateRendering(t *testing.T) {
	customTemplates := map[string]string{
		"log_summary": "🔍 日志文件: {{.FilePath}}\n时间: {{.Time}}\n内容: {{.Summary}}",
	}

	data := TemplateData{
		FilePath: "/var/log/test.log",
		Time:     "2023-01-01 12:00:00",
		Summary:  "Test summary",
	}

	result, err := RenderTemplate("log_summary", data, customTemplates, getTelegramDefaultTemplates)
	assert.NoError(t, err)

	expected := "🔍 日志文件: /var/log/test.log\n时间: 2023-01-01 12:00:00\n内容: Test summary"
	assert.Equal(t, expected, result)
}

func TestTemplateRenderingFallback(t *testing.T) {
	// Test with empty templates - should use defaults
	data := TemplateData{
		FilePath: "/var/log/test.log",
		Time:     "2023-01-01 12:00:00",
		Summary:  "Test summary",
	}

	result, err := RenderTemplate("log_summary", data, map[string]string{}, getTelegramDefaultTemplates)
	assert.NoError(t, err)
	assert.Contains(t, result, "Log Summary Notification")
	assert.Contains(t, result, "/var/log/test.log")
	assert.Contains(t, result, "Test summary")
}

func TestTemplateRenderingWithInvalidSyntax(t *testing.T) {
	invalidTemplates := map[string]string{
		"log_summary": "Invalid template {{.FilePath", // Missing closing brace
	}

	data := TemplateData{
		FilePath: "/var/log/test.log",
		Time:     "2023-01-01 12:00:00",
		Summary:  "Test summary",
	}

	// Should fallback to default template when parsing fails
	result, err := RenderTemplate("log_summary", data, invalidTemplates, getTelegramDefaultTemplates)
	assert.NoError(t, err) // Should not error because it falls back to default
	assert.Contains(t, result, "Log Summary Notification")
}

func TestGetDefaultTemplates(t *testing.T) {
	templates := getTelegramDefaultTemplates()

	assert.NotEmpty(t, templates["log_summary"])

	// Test that templates contain expected elements
	assert.Contains(t, templates["log_summary"], "{{.FilePath}}")
	assert.Contains(t, templates["log_summary"], "{{.Time}}")
	assert.Contains(t, templates["log_summary"], "{{.Summary}}")
}

func TestMessageTemplatesGetTemplateMap(t *testing.T) {
	mt := config.TelegramMessageTemplates{
		LogSummary: "Custom log template: {{.Summary}}",
	}

	templateMap := mt.GetTemplateMap()

	assert.Equal(t, "Custom log template: {{.Summary}}", templateMap["log_summary"])
}

func TestRenderTemplate(t *testing.T) {
	customTemplates := map[string]string{
		"log_summary": "Custom: {{.Summary}} from {{.FilePath}}",
	}

	data := TemplateData{
		FilePath: "/var/log/test.log",
		Time:     "2023-01-01 12:00:00",
		Summary:  "Test summary",
	}

	result, err := RenderTemplate("log_summary", data, customTemplates, getTelegramDefaultTemplates)
	assert.NoError(t, err)
	assert.Equal(t, "Custom: Test summary from /var/log/test.log", result)
}

func TestRenderTemplate_FallbackToDefault(t *testing.T) {
	customTemplates := map[string]string{} // Empty templates

	data := TemplateData{
		FilePath: "/var/log/test.log",
		Time:     "2023-01-01 12:00:00",
		Summary:  "Test summary",
	}

	result, err := RenderTemplate("log_summary", data, customTemplates, getTelegramDefaultTemplates)
	assert.NoError(t, err)
	assert.Contains(t, result, "Log Summary Notification")
}

func TestRenderTemplate_InvalidTemplateFallback(t *testing.T) {
	invalidTemplates := map[string]string{
		"log_summary": "Invalid template {{.FilePath", // Missing closing brace
	}

	data := TemplateData{
		FilePath: "/var/log/test.log",
		Time:     "2023-01-01 12:00:00",
		Summary:  "Test summary",
	}

	result, err := RenderTemplate("log_summary", data, invalidTemplates, getTelegramDefaultTemplates)
	assert.NoError(t, err) // Should fallback to default template
	assert.Contains(t, result, "Log Summary Notification")
}

func TestSendLogSummary_Common(t *testing.T) {
	// Test the common SendLogSummary function with a mock notifier
	mockNotifier := &mockNotifier{}
	
	filePath := "/var/log/test.log"
	summary := "Test summary content"
	
	err := SendLogSummary(mockNotifier, filePath, summary, map[string]string{}, func() map[string]string {
		return map[string]string{
			"log_summary": "Mock: {{.Summary}} from {{.FilePath}}",
		}
	})
	
	assert.NoError(t, err)
	assert.True(t, mockNotifier.sendMessageCalled)
	assert.Contains(t, mockNotifier.lastMessage, "Mock: Test summary content from /var/log/test.log")
}

func TestCreateNotifiers(t *testing.T) {
	cfg := &config.Config{
		OpenAI: config.OpenAIConfig{
			APIKey: "test-key",
		},
		Telegram: config.TelegramConfig{
			BotToken: "test-token",
			ChatID:   "-100123",
		},
		Email: config.EmailConfig{
			SMTPHost:  "smtp.gmail.com",
			SMTPPort:  587,
			Username:  "test@gmail.com",
			Password:  "test-password",
			FromEmail: "test@gmail.com",
			ToEmails:  []string{"recipient@gmail.com"},
			Subject:   "Test Subject",
			UseTLS:    true,
		},
		ChatID: "-100123",
	}

	// Test with both notifiers enabled
	notifiers, err := CreateNotifiers(cfg, []string{"telegram", "email"})
	assert.NoError(t, err)
	assert.Len(t, notifiers, 2)

	// Test with only telegram
	notifiers, err = CreateNotifiers(cfg, []string{"telegram"})
	assert.NoError(t, err)
	assert.Len(t, notifiers, 1)
	assert.IsType(t, &TelegramNotifier{}, notifiers[0])

	// Test with only email
	notifiers, err = CreateNotifiers(cfg, []string{"email"})
	assert.NoError(t, err)
	assert.Len(t, notifiers, 1)
	assert.IsType(t, &EmailNotifier{}, notifiers[0])

	// Test with no notifiers specified (should enable all configured)
	notifiers, err = CreateNotifiers(cfg, []string{})
	assert.NoError(t, err)
	assert.Len(t, notifiers, 2)

	// Test with invalid configuration
	invalidCfg := &config.Config{
		OpenAI: config.OpenAIConfig{
			APIKey: "test-key",
		},
		Telegram: config.TelegramConfig{
			BotToken: "", // Invalid
			ChatID:   "-100123",
		},
		Email: config.EmailConfig{
			SMTPHost: "", // Invalid
		},
		ChatID: "-100123",
	}

	notifiers, err = CreateNotifiers(invalidCfg, []string{})
	assert.Error(t, err)
	assert.Nil(t, notifiers)
}

// mockNotifier is a mock implementation of the Notifier interface for testing
type mockNotifier struct {
	sendMessageCalled bool
	lastMessage      string
}

func (m *mockNotifier) SendMessage(message string) error {
	m.sendMessageCalled = true
	m.lastMessage = message
	return nil
}

func (m *mockNotifier) SendLogSummary(filePath, summary string) error {
	m.sendMessageCalled = true
	m.lastMessage = "Mock log summary"
	return nil
}
