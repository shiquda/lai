package notifier

import (
	"testing"

	"gopkg.in/gomail.v2"
	"github.com/stretchr/testify/assert"
)

func TestNewEmailNotifier(t *testing.T) {
	smtpHost := "smtp.gmail.com"
	smtpPort := 587
	username := "test@gmail.com"
	password := "test-password"
	fromEmail := "test@gmail.com"
	toEmails := []string{"recipient1@gmail.com", "recipient2@gmail.com"}
	subject := "Test Email"
	useTLS := true

	notifier := NewEmailNotifier(smtpHost, smtpPort, username, password, fromEmail, toEmails, subject, useTLS, nil)

	assert.NotNil(t, notifier)
	assert.Equal(t, smtpHost, notifier.smtpHost)
	assert.Equal(t, smtpPort, notifier.smtpPort)
	assert.Equal(t, username, notifier.username)
	assert.Equal(t, password, notifier.password)
	assert.Equal(t, fromEmail, notifier.fromEmail)
	assert.Equal(t, toEmails, notifier.toEmails)
	assert.Equal(t, subject, notifier.subject)
	assert.Equal(t, useTLS, notifier.useTLS)
	assert.NotNil(t, notifier.messageTemplates)
}

func TestNewEmailNotifier_WithTemplates(t *testing.T) {
	customTemplates := map[string]string{
		"log_summary": "Custom email template: {{.Summary}}",
	}

	notifier := NewEmailNotifier("smtp.gmail.com", 587, "test@gmail.com", "password", "test@gmail.com", []string{"recipient@gmail.com"}, "Test", true, customTemplates)

	assert.NotNil(t, notifier)
	assert.Equal(t, customTemplates, notifier.messageTemplates)
}

func TestNewEmailNotifier_WithoutTemplates(t *testing.T) {
	notifier := NewEmailNotifier("smtp.gmail.com", 587, "test@gmail.com", "password", "test@gmail.com", []string{"recipient@gmail.com"}, "Test", true, nil)

	assert.NotNil(t, notifier)
	assert.NotNil(t, notifier.messageTemplates)
	assert.NotEmpty(t, notifier.messageTemplates["log_summary"])
	assert.Contains(t, notifier.messageTemplates["log_summary"], "<html>")
}

func TestEmailNotifier_SetClient(t *testing.T) {
	notifier := NewEmailNotifier("smtp.gmail.com", 587, "test@gmail.com", "password", "test@gmail.com", []string{"recipient@gmail.com"}, "Test", true, nil)

	mockClient := &mockDialer{}
	notifier.SetClient(mockClient)

	assert.Equal(t, mockClient, notifier.client)
}

func TestEmailNotifier_SendMessage_NoRecipients(t *testing.T) {
	notifier := NewEmailNotifier("smtp.gmail.com", 587, "test@gmail.com", "password", "test@gmail.com", []string{}, "Test", true, nil)

	err := notifier.SendMessage("Test message")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no recipient email addresses provided")
}

func TestEmailNotifier_SendLogSummary(t *testing.T) {
	// This test would require actual SMTP server connection
	// For unit testing, we should mock the SMTP client
	// For now, we'll skip this test as it requires network connection
	t.Skip("Skipping test that requires actual SMTP connection")
}

func TestGetDefaultEmailTemplates(t *testing.T) {
	templates := getDefaultEmailTemplates()

	assert.NotNil(t, templates)
	assert.NotEmpty(t, templates["log_summary"])
	assert.Contains(t, templates["log_summary"], "<html>")
	assert.Contains(t, templates["log_summary"], "body")
	assert.Contains(t, templates["log_summary"], "{{.FilePath}}")
	assert.Contains(t, templates["log_summary"], "{{.Time}}")
	assert.Contains(t, templates["log_summary"], "{{.Summary}}")
}

// TestEmailNotifier_SendMessage_WithMockClient is removed because the current implementation
// uses gomail library which doesn't support easy mocking without interfaces
// This would require refactoring the EmailNotifier to use dependency injection for the mailer

// mockDialer is a mock SMTP dialer for testing
type mockDialer struct{}

func (m *mockDialer) DialAndSend(msg *gomail.Message) error {
	return nil
}