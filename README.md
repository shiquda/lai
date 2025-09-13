# Lai: AI-Powered Log Monitoring

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/doc/install)
[![AI Powered](https://img.shields.io/badge/AI-Powered-brightgreen.svg)]()
[![License](https://img.shields.io/badge/License-AGPL--3.0-yellow.svg)](LICENSE)

Stop manually checking logs. Let AI watch, analyze, and notify you when something important happens.

> **Note**: This project is under active development. Contributions welcome!

## 🚀 5-Minute Quick Start

### 1. Install Lai

```bash
# Download latest release (Linux)
wget https://github.com/shiquda/lai/releases/latest/download/lai-v*-linux-amd64
mkdir -p ~/.local/bin && mv lai-v*-linux-amd64 ~/.local/bin/lai
chmod +x ~/.local/bin/lai
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc

# Or build from source
git clone https://github.com/shiquda/lai.git && cd lai
make build && cp lai ~/.local/bin/
```

### 2. Configure Notifications

```bash
# Set up OpenAI for AI analysis
lai config set notifications.openai.api_key "sk-your-key"

# Configure Telegram (recommended)
lai config set notifications.providers.telegram.enabled true
lai config set notifications.providers.telegram.bot_token "123456:ABC-DEF"
lai config set notifications.providers.telegram.chat_id "-100123456789"

# Or configure Email
lai config set notifications.providers.email.enabled true
lai config set notifications.providers.email.smtp_host "smtp.gmail.com"
lai config set notifications.providers.email.smtp_port "587"
# ... more email config
```

### 3. Start Monitoring

```bash
# Monitor application logs
lai monitor file /var/log/app.log

# Monitor Docker containers
lai monitor command "docker logs webapp -f"

# Run as daemon
lai monitor file /var/log/nginx/error.log -d -n "nginx-monitor"
```

## ✨ Key Features

- **🤖 AI-Powered Analysis**: LLMs automatically summarize log changes and identify issues
- **📱 Smart Notifications**: Get alerts via Telegram, Email, Discord, or Slack
- **🔄 Universal Monitoring**: Watch any log file or command output
- **🎨 Colored Output**: Distinguish stdout/stderr with configurable colors in exec mode
- **🔌 Zero Integration**: Works with any existing application - no code changes needed
- **⚡ Real-time Processing**: Instant analysis and notification delivery

## 📖 Use Cases

### Application Monitoring
```bash
# Background monitoring with custom name
lai monitor file /var/log/nginx/error.log -d -n "nginx-errors"
lai list  # View running monitors
lai logs nginx-errors -f  # Check monitor logs
```

### Docker Container Monitoring
```bash
# Monitor specific container
lai monitor command "docker logs webapp -f" -d -n "webapp-monitor"

# Monitor with custom thresholds
lai monitor command "docker logs db -f" --line-threshold 5 --interval 10s
```

### Build/CI Process Monitoring
```bash
# Get summary when build completes
lai monitor command "npm run build" --final-summary

# Monitor tests with error detection
lai monitor command "npm test" -l 3 -i 15s

# Monitor command output with colored display (stdout: gray, stderr: red)
lai exec "npm run build" --final-summary

# Monitor long-running processes
lai exec "python train_model.py" -d -n "model-training"
```

## 🔧 Configuration Options

### Interactive Setup (Recommended)
```bash
lai config interactive  # Guided configuration interface
```

### Command Line Configuration
```bash
# View current configuration
lai config list

# Set OpenAI configuration
lai config set notifications.openai.api_key "sk-your-key"
lai config set notifications.openai.model "gpt-3.5-turbo"

# Configure notification providers
lai config set notifications.providers.telegram.enabled true
lai config set notifications.providers.telegram.bot_token "your-token"
lai config set notifications.providers.telegram.chat_id "your-chat-id"

# Set monitoring preferences
lai config set defaults.line_threshold 10
lai config set defaults.check_interval "30s"
lai config set defaults.language "English"

# Configure colored output for exec command
lai config set display.colors.enabled true
lai config set display.colors.stdout "gray"
lai config set display.colors.stderr "red"

# Reset configuration to defaults
lai config reset
```

### Process Management
```bash
lai list           # Show all running monitors
lai stop <name>    # Stop a monitor
lai resume <name>  # Restart a stopped monitor
lai clean          # Remove stopped entries
```

## 📚 Advanced Topics

### Supported Notification Providers

- **Telegram**: Bot token + chat ID
- **Email**: SMTP configuration with multiple providers (SendGrid, Gmail, etc.)
- **Discord**: Bot token or webhook
- **Slack**: Webhook or OAuth token
- **Pushover**: Mobile notifications
- **Twilio**: SMS alerts
- **PagerDuty**: Incident management
- **DingTalk/WeChat**: Chinese platforms

### Configuration File

The global configuration is stored at `~/.lai/config.yaml`. You can edit this file directly or use the `lai config` commands.

### Advanced Features

- **Error-only mode**: Only notify on errors/exceptions
- **Final summary**: Get summary when monitoring stops
- **Custom thresholds**: Adjust sensitivity and check intervals
- **Multi-language AI responses**: Configure response language
- **Daemon mode**: Run monitoring processes in background

## 🛠️ Development

### Building from Source
```bash
git clone https://github.com/shiquda/lai.git
cd lai
make build        # Build the application
make test-quick   # Run tests
make test         # Run full test suite with coverage
```

### Contributing
1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass: `make test`
5. Submit a pull request

## 📋 Roadmap

### Recently Completed ✅
- [x] Unified monitoring interface with single `monitor` command
- [x] Interactive configuration TUI
- [x] Multi-provider notification system (Telegram, Email, Discord, Slack)
- [x] Cross-platform improvements
- [x] Configuration validation and metadata system

### Upcoming Features 🚀
- [ ] Webhook notifications support
- [ ] Advanced log filtering and pattern matching
- [ ] Integration with monitoring tools (Prometheus, Grafana)

## 📄 License

AGPL-3.0 - see LICENSE file for details.

---

**[Documentation](docs/)** | **[Configuration Reference](docs/CONFIGURATION.md)** | **[Architecture](docs/ARCHITECTURE.md)**