package notifier

import (
	"fmt"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"gopkg.in/gomail.v2"
)

// EmailNotifier implements the Notifier interface for email notifications.
// It uses the gomail library to send emails via SMTP servers.
type EmailNotifier struct {
	smtpHost         string            // SMTP server hostname (e.g., "smtp.gmail.com")
	smtpPort         int               // SMTP server port (e.g., 587 for TLS, 465 for SSL)
	username         string            // SMTP authentication username
	password         string            // SMTP authentication password
	fromEmail        string            // Sender email address
	toEmails         []string          // List of recipient email addresses
	subject          string            // Email subject line
	useTLS           bool              // Whether to use TLS encryption
	client           interface{}       // Placeholder for SMTP client injection (for testing)
	messageTemplates map[string]string // Custom email templates
}

// NewEmailNotifier creates a new EmailNotifier instance with the provided configuration.
// It sets up the SMTP client configuration and message templates.
//
// Parameters:
//   - smtpHost: SMTP server hostname (e.g., "smtp.gmail.com")
//   - smtpPort: SMTP server port (e.g., 587 for TLS, 465 for SSL)
//   - username: SMTP authentication username
//   - password: SMTP authentication password (use app passwords for Gmail)
//   - fromEmail: Sender email address
//   - toEmails: List of recipient email addresses
//   - subject: Email subject line
//   - useTLS: Whether to use TLS encryption
//   - templates: Custom email templates (nil to use defaults)
//
// Returns:
//   - Configured EmailNotifier instance
func NewEmailNotifier(smtpHost string, smtpPort int, username, password, fromEmail string, toEmails []string, subject string, useTLS bool, templates map[string]string) *EmailNotifier {
	if templates == nil {
		templates = getDefaultEmailTemplates()
	}
	return &EmailNotifier{
		smtpHost:         smtpHost,
		smtpPort:         smtpPort,
		username:         username,
		password:         password,
		fromEmail:        fromEmail,
		toEmails:         toEmails,
		subject:          subject,
		useTLS:           useTLS,
		messageTemplates: templates,
	}
}

// SetClient sets an SMTP client for dependency injection.
// This is primarily used for testing purposes to mock the SMTP client.
//
// Parameters:
//   - client: Mock SMTP client implementing the DialAndSend method
func (e *EmailNotifier) SetClient(client interface{}) {
	e.client = client
}

// SendMessage sends an email with the given message body.
// The message should be formatted as HTML for proper email rendering.
//
// Parameters:
//   - message: HTML formatted message content
//
// Returns:
//   - Error if email sending fails (authentication, connection, etc.)
func (e *EmailNotifier) SendMessage(message string) error {
	if len(e.toEmails) == 0 {
		return fmt.Errorf("no recipient email addresses provided")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", e.fromEmail)
	m.SetHeader("To", e.toEmails...)
	m.SetHeader("Subject", e.subject)
	m.SetBody("text/html", message)

	d := gomail.NewDialer(e.smtpHost, e.smtpPort, e.username, e.password)

	// For Gmail port 587, TLS should be enabled by default
	// Only disable TLS if explicitly set to false
	if !e.useTLS {
		d.SSL = false
		d.TLSConfig = nil
	}

	err := d.DialAndSend(m)
	if err != nil {
		return fmt.Errorf("SMTP error: %w", err)
	}
	return nil
}

// SendLogSummary sends a log summary email using templates.
// This method converts Markdown summary to HTML before sending.
//
// Parameters:
//   - filePath: Path to the log file
//   - summary: AI-generated log summary content (Markdown format)
//
// Returns:
//   - Error if template rendering or email sending fails
func (e *EmailNotifier) SendLogSummary(filePath, summary string) error {
	// Convert Markdown to HTML for better email rendering
	htmlSummary := convertMarkdownToHTML(summary)
	return e.sendLogSummaryWithHTML(filePath, htmlSummary)
}

// sendLogSummaryWithHTML sends a log summary email using templates with HTML content.
// This method uses a custom template rendering to avoid HTML escaping.
//
// Parameters:
//   - filePath: Path to the log file
//   - htmlSummary: HTML-formatted log summary content
//
// Returns:
//   - Error if template rendering or email sending fails
func (e *EmailNotifier) sendLogSummaryWithHTML(filePath, htmlSummary string) error {
	// Get template
	templateStr := e.messageTemplates["log_summary"]
	if templateStr == "" {
		// Fall back to default template
		defaultTemplates := getDefaultEmailTemplates()
		templateStr = defaultTemplates["log_summary"]
	}

	// Create template data
	data := TemplateData{
		FilePath: filePath,
		Time:     time.Now().Format("2006-01-02 15:04:05"),
		Summary:  htmlSummary, // This is already HTML
	}

	// Custom template rendering to avoid HTML escaping for the summary field
	message, err := e.renderEmailTemplate(templateStr, data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	return e.SendMessage(message)
}

// renderEmailTemplate renders an email template with proper HTML handling.
// It avoids HTML escaping for the summary field while keeping other fields safe.
//
// Parameters:
//   - templateStr: Template string
//   - data: Template data
//
// Returns:
//   - Rendered HTML message
//   - Error if rendering fails
func (e *EmailNotifier) renderEmailTemplate(templateStr string, data TemplateData) (string, error) {
	// Replace placeholders manually to avoid HTML escaping
	message := templateStr

	// Replace non-HTML fields (these are safe)
	message = strings.ReplaceAll(message, "{{.FilePath}}", data.FilePath)
	message = strings.ReplaceAll(message, "{{.Time}}", data.Time)

	// Replace HTML field without escaping
	message = strings.ReplaceAll(message, "{{.Summary}}", data.Summary)

	return message, nil
}

// convertMarkdownToHTML converts Markdown text to safe HTML.
// It uses blackfriday for Markdown parsing and bluemonday for HTML sanitization.
//
// Parameters:
//   - markdown: Markdown formatted text
//
// Returns:
//   - Safe HTML string
func convertMarkdownToHTML(markdown string) string {
	// Convert Markdown to HTML with extensions for better formatting
	unsafeHTML := blackfriday.Run([]byte(markdown), blackfriday.WithExtensions(
		blackfriday.CommonExtensions|blackfriday.AutoHeadingIDs|blackfriday.Tables|blackfriday.FencedCode,
	))

	// Sanitize HTML to prevent XSS attacks
	p := bluemonday.UGCPolicy()
	// Allow common formatting elements
	p.AllowStandardURLs()
	p.AllowStandardAttributes()
	p.AllowElements("h1", "h2", "h3", "h4", "h5", "h6", "p", "br", "hr")
	p.AllowElements("ul", "ol", "li", "dl", "dt", "dd")
	p.AllowElements("strong", "em", "b", "i", "u", "del", "ins")
	p.AllowElements("blockquote", "pre", "code", "samp", "kbd")
	p.AllowElements("a", "img")
	p.AllowElements("table", "thead", "tbody", "tr", "th", "td")

	// Allow attributes for styling
	p.AllowAttrs("class").Matching(bluemonday.SpaceSeparatedTokens).OnElements("code", "pre")
	p.AllowAttrs("style").OnElements("p", "div", "span", "h1", "h2", "h3", "h4", "h5", "h6")

	// Sanitize the HTML
	safeHTML := p.SanitizeBytes(unsafeHTML)

	return string(safeHTML)
}

// ConvertMarkdownToHTML is a public wrapper for convertMarkdownToHTML.
// This allows external testing and reuse of the conversion functionality.
//
// Parameters:
//   - markdown: Markdown formatted text
//
// Returns:
//   - Safe HTML string
func ConvertMarkdownToHTML(markdown string) string {
	return convertMarkdownToHTML(markdown)
}

// Note: The following TODO methods were considered but implemented using the common template system:
// - buildEmailMessage: Handled by gomail.Message construction
// - renderTemplate: Handled by the common RenderTemplate function in notifier.go

// getDefaultEmailTemplates returns default email templates in HTML format.
// These templates are used when no custom templates are configured.
// The summary content is automatically converted from Markdown to HTML.
//
// Returns:
//   - Map of template names to template content
func getDefaultEmailTemplates() map[string]string {
	return map[string]string{
		"log_summary": `<html>
<head>
	<meta charset="utf-8">
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif; padding: 20px; line-height: 1.6; }
		.header { background-color: #f5f5f5; padding: 15px; border-radius: 5px; margin-bottom: 20px; }
		.title { color: #D32F2F; margin: 0 0 10px 0; }
		.meta { color: #666; font-size: 14px; margin: 5px 0; }
		.content { border-left: 4px solid #D32F2F; padding-left: 15px; margin: 20px 0; }
		h1, h2, h3, h4, h5, h6 { color: #333; margin-top: 20px; }
		code { background-color: #f4f4f4; padding: 2px 4px; border-radius: 3px; font-family: 'Courier New', monospace; }
		pre { background-color: #f4f4f4; padding: 10px; border-radius: 5px; overflow-x: auto; }
		blockquote { border-left: 4px solid #ddd; margin: 0; padding-left: 15px; color: #666; }
		table { border-collapse: collapse; width: 100%; margin: 10px 0; }
		th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
		th { background-color: #f2f2f2; }
	</style>
</head>
<body>
	<div class="header">
		<h2 class="title">📋 Log Summary Notification</h2>
		<p class="meta"><strong>File:</strong> {{.FilePath}}</p>
		<p class="meta"><strong>Time:</strong> {{.Time}}</p>
	</div>
	<div class="content">
		{{.Summary}}
	</div>
	<hr style="margin-top: 30px;">
	<p style="color: #999; font-size: 12px;">Generated by Lai - AI Log Monitoring Tool</p>
</body>
</html>`,
	}
}
