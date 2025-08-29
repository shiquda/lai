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

	notifier := NewTelegramNotifier("test-token", "-100123456789", customTemplates)

	data := TemplateData{
		FilePath: "/var/log/test.log",
		Time:     "2023-01-01 12:00:00",
		Summary:  "Test summary",
	}

	result, err := notifier.renderTemplate("log_summary", data)
	assert.NoError(t, err)

	expected := "🔍 日志文件: /var/log/test.log\n时间: 2023-01-01 12:00:00\n内容: Test summary"
	assert.Equal(t, expected, result)
}

func TestTemplateRenderingFallback(t *testing.T) {
	// Test with empty templates - should use defaults
	notifier := NewTelegramNotifier("test-token", "-100123456789", map[string]string{})

	data := TemplateData{
		FilePath: "/var/log/test.log",
		Time:     "2023-01-01 12:00:00",
		Summary:  "Test summary",
	}

	result, err := notifier.renderTemplate("log_summary", data)
	assert.NoError(t, err)
	assert.Contains(t, result, "Log Summary Notification")
	assert.Contains(t, result, "/var/log/test.log")
	assert.Contains(t, result, "Test summary")
}

func TestTemplateRenderingWithInvalidSyntax(t *testing.T) {
	invalidTemplates := map[string]string{
		"log_summary": "Invalid template {{.FilePath", // Missing closing brace
	}

	notifier := NewTelegramNotifier("test-token", "-100123456789", invalidTemplates)

	data := TemplateData{
		FilePath: "/var/log/test.log",
		Time:     "2023-01-01 12:00:00",
		Summary:  "Test summary",
	}

	// Should fallback to default template when parsing fails
	result, err := notifier.renderTemplate("log_summary", data)
	assert.NoError(t, err) // Should not error because it falls back to default
	assert.Contains(t, result, "Log Summary Notification")
}

func TestGetDefaultTemplates(t *testing.T) {
	templates := getDefaultTemplates()

	assert.NotEmpty(t, templates["log_summary"])

	// Test that templates contain expected elements
	assert.Contains(t, templates["log_summary"], "{{.FilePath}}")
	assert.Contains(t, templates["log_summary"], "{{.Time}}")
	assert.Contains(t, templates["log_summary"], "{{.Summary}}")
}

func TestMessageTemplatesGetTemplateMap(t *testing.T) {
	mt := config.MessageTemplates{
		LogSummary: "Custom log template: {{.Summary}}",
	}

	templateMap := mt.GetTemplateMap()

	assert.Equal(t, "Custom log template: {{.Summary}}", templateMap["log_summary"])
}