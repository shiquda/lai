# Email Notifications

Lai supports email notifications alongside Telegram, providing flexible notification options for your log monitoring needs.

## 📧 Configuration

### Basic SMTP Setup

```bash
# Configure SMTP server settings
lai config set notifications.email.smtp_host "smtp.gmail.com"
lai config set notifications.email.smtp_port "587"
lai config set notifications.email.use_tls "true"

# Configure authentication
lai config set notifications.email.username "your-email@gmail.com"
lai config set notifications.email.password "your-app-password"
lai config set notifications.email.from_email "your-email@gmail.com"

# Configure recipients
lai config set notifications.email.to_emails '["admin@company.com", "devops@company.com"]'
lai config set notifications.email.subject "🚨 Lai Log Alert"
```

### Email Service Providers

#### Gmail
```bash
# Gmail SMTP configuration
lai config set notifications.email.smtp_host "smtp.gmail.com"
lai config set notifications.email.smtp_port "587"
lai config set notifications.email.use_tls "true"
# Note: Use App Password, not your regular password
```

#### Outlook/Office 365
```bash
# Outlook SMTP configuration
lai config set notifications.email.smtp_host "smtp.office365.com"
lai config set notifications.email.smtp_port "587"
lai config set notifications.email.use_tls "true"
```

#### Yahoo Mail
```bash
# Yahoo SMTP configuration
lai config set notifications.email.smtp_host "smtp.mail.yahoo.com"
lai config set notifications.email.smtp_port "587"
lai config set notifications.email.use_tls "true"
```

## 🎨 Custom Email Templates

You can customize the email notification format using HTML templates:

```bash
# Set custom email template
lai config set notifications.email.message_templates.log_summary '
<html>
<body style="font-family: Arial, sans-serif; padding: 20px; background-color: #f5f5f5;">
  <div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1);">
    <h1 style="color: #d32f2f; margin-bottom: 20px;">🚨 Log Alert</h1>
    <div style="background-color: #f8f9fa; padding: 15px; border-radius: 5px; margin-bottom: 20px;">
      <p><strong>File:</strong> {{.FilePath}}</p>
      <p><strong>Time:</strong> {{.Time}}</p>
      <p><strong>Process:</strong> {{.ProcessName}}</p>
      <p><strong>Lines:</strong> {{.LineCount}}</p>
    </div>
    <div style="border-left: 4px solid #d32f2f; padding-left: 15px;">
      <h3>Summary:</h3>
      <pre style="white-space: pre-wrap; font-family: monospace;">{{.Summary}}</pre>
    </div>
  </div>
</body>
</html>
'
```

### Template Variables

- `{{.FilePath}}` - Path to the log file
- `{{.Time}}` - Timestamp when the notification was sent
- `{{.Summary}}` - AI-generated log summary
- `{{.ProcessName}}` - Name of the monitoring process (if set)
- `{{.LineCount}}` - Number of lines that triggered the notification

## 🔧 Usage Examples

### Basic Email Notification
```bash
# Send notifications only to email
lai start /var/log/app.log --notifiers email
```

### Combined Notifications
```bash
# Send to both Telegram and Email
lai start /var/log/app.log --notifiers telegram,email
```

### Email-Only Monitoring
```bash
# Background monitoring with email notifications
lai start /var/log/app.log -d -n "app-monitor" --notifiers email
```

### Error-Only Email Alerts
```bash
# Send email only when errors are detected
lai start /var/log/app.log --notifiers email --error-only
```

## 🔒 Security Considerations

### App Passwords
For Gmail and other email providers, use **App Passwords** instead of your regular password:

1. Go to your Google Account settings
2. Enable 2-Step Verification
3. Generate an App Password
4. Use the app password in Lai configuration

### TLS Encryption
Lai uses TLS encryption by default for email sending. Ensure your email provider supports TLS:

```bash
# Enable TLS (default)
lai config set notifications.email.use_tls "true"

# Disable TLS for local testing (not recommended for production)
lai config set notifications.email.use_tls "false"
```

## 🛠️ Troubleshooting

### Common Issues

#### Authentication Failed
```bash
# Check if username and password are correct
lai config list notifications.email

# Ensure you're using app passwords for Gmail
# Regular Gmail passwords won't work with SMTP
```

#### Connection Timeout
```bash
# Check SMTP server and port
lai config list notifications.email.smtp_host
lai config list notifications.email.smtp_port

# Verify firewall settings
# Try using port 587 with TLS or 465 with SSL
```

#### Emails Not Delivered
```bash
# Check recipient email addresses
lai config list notifications.email.to_emails

# Verify sender email is configured
lai config list notifications.email.from_email

# Check spam/junk folder
```

### Testing Email Configuration

```bash
# Create a test log file
echo "Test log entry" > /tmp/test.log

# Start monitoring with email only
lai start /tmp/test.log --notifiers email --line-threshold 1

# Trigger notification by adding more content
echo "Another log entry" >> /tmp/test.log
```

## 📊 Advanced Configuration

### Multiple Email Recipients
```bash
# Set multiple recipients
lai config set notifications.email.to_emails '["team@company.com", "alerts@company.com", "admin@company.com"]'
```

### Custom Email Subject
```bash
# Dynamic subject with file name
lai config set notifications.email.subject "🚨 Alert from {{.ProcessName}} - {{.FilePath}}"
```

### Conditional Email Notifications
```bash
# Send email only for specific files
lai start /var/log/critical.log --notifiers email
lai start /var/log/normal.log --notifiers telegram
```

## 🔧 Integration with CI/CD

### GitHub Actions Example
```yaml
name: Log Monitoring
on:
  push:
    paths: 
      - 'logs/**'

jobs:
  monitor:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Lai
        run: |
          wget https://github.com/shiquda/lai/releases/latest/download/lai-v*-linux-amd64
          chmod +x lai
          sudo mv lai /usr/local/bin/
      
      - name: Configure Email
        run: |
          lai config set notifications.openai.api_key ${{ secrets.OPENAI_API_KEY }}
          lai config set notifications.email.smtp_host "smtp.gmail.com"
          lai config set notifications.email.smtp_port "587"
          lai config set notifications.email.username ${{ secrets.EMAIL_USERNAME }}
          lai config set notifications.email.password ${{ secrets.EMAIL_PASSWORD }}
          lai config set notifications.email.from_email ${{ secrets.EMAIL_FROM }}
          lai config set notifications.email.to_emails '["devops@company.com"]'
      
      - name: Monitor Logs
        run: |
          lai start logs/app.log --notifiers email --line-threshold 5 &
```

## 📝 Example Configurations

### Production Setup
```yaml
notifications:
  email:
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    use_tls: true
    username: "alerts@company.com"
    password: "your-app-password"
    from_email: "alerts@company.com"
    to_emails: 
      - "devops@company.com"
      - "admin@company.com"
    subject: "🚨 Production Alert - {{.ProcessName}}"
    message_templates:
      log_summary: |
        <html>
        <body style="font-family: Arial, sans-serif; padding: 20px;">
          <div style="background-color: #d32f2f; color: white; padding: 10px; border-radius: 5px;">
            <h2>🚨 PRODUCTION ALERT</h2>
          </div>
          <div style="padding: 20px; border: 1px solid #ddd; border-radius: 5px;">
            <p><strong>File:</strong> {{.FilePath}}</p>
            <p><strong>Time:</strong> {{.Time}}</p>
            <p><strong>Process:</strong> {{.ProcessName}}</p>
            <hr>
            <h3>Summary:</h3>
            <pre>{{.Summary}}</pre>
          </div>
        </body>
        </html>
```

### Development Setup
```yaml
notifications:
  email:
    smtp_host: "smtp.gmail.com"
    smtp_port: 587
    use_tls: true
    username: "dev@company.com"
    password: "your-app-password"
    from_email: "dev@company.com"
    to_emails: 
      - "dev-team@company.com"
    subject: "🔍 Dev Log Summary - {{.ProcessName}}"
```

## 🤝 Contributing

To add new email providers or improve email functionality:

1. Update the `EmailConfig` struct in `internal/config/config.go`
2. Modify the `EmailNotifier` implementation in `internal/notifier/email.go`
3. Add tests in `internal/notifier/email_test.go`
4. Update documentation in this file

## 📚 Related Documentation

- [Configuration Reference](CONFIGURATION.md) - Complete configuration options
- [Architecture Overview](ARCHITECTURE.md) - Understanding the notification system
- [Development Guide](DEVELOPMENT.md) - Building and testing