# Lai: AI-Powered Log Monitoring

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/doc/install)
[![AI Powered](https://img.shields.io/badge/AI-Powered-brightgreen.svg)]()
[![Telegram](https://img.shields.io/badge/Notifications-Telegram-blue.svg)](https://telegram.org/)
[![Email](https://img.shields.io/badge/Notifications-Email-red.svg)](mailto:)
[![License](https://img.shields.io/badge/License-AGPL--3.0-yellow.svg)](LICENSE)

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/shiquda/lai)

Stop manually checking logs. Let AI watch, analyze, and notify you when something important happens.

> Note: This project is under active development. Any kind of contributions are welcome!

## ✨ Core Features

- **🤖 AI-Powered Analysis**: LLMs automatically summarizes log changes and identifies critical issues
- **📱 Instant Notifications**: Get smart alerts via Telegram, Email, Discord, or Slack when errors or important events occur
- **🔄 Universal Monitoring**: Watch any log file or command output (Docker logs, application output, build processes)
- **🔌 Hot-Pluggable**: No code changes required - works with any existing project or application
- **📝 Customizable Templates**: Personalize notification messages with custom templates for each channel

## ⚡ Installation

### Option 1: Install from Release (Recommended)

```bash
# Download latest release for your platform
# For Linux:
wget https://github.com/shiquda/lai/releases/latest/download/lai-v*-linux-amd64
# For macOS Intel:
# wget https://github.com/shiquda/lai/releases/latest/download/lai-v*-darwin-amd64
# For macOS Apple Silicon:
# wget https://github.com/shiquda/lai/releases/latest/download/lai-v*-darwin-arm64

# Create local bin directory and rename binary
mkdir -p ~/.local/bin
mv lai-v*-linux-amd64 ~/.local/bin/lai

# Make executable
chmod +x ~/.local/bin/lai

# Add to PATH if needed (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Verify installation
lai version
```

### Option 2: Build from Source

**Prerequisites:**

```bash
# Install Go 1.21+ if not installed
sudo apt update
sudo apt install golang-go

# Verify Go version
go version  # Should be 1.21 or higher
```

**Build and install:**

```bash
# Clone repository
git clone https://github.com/shiquda/lai.git
cd lai

# Build the application
make build

# Install to ~/.local/bin
mkdir -p ~/.local/bin
cp lai ~/.local/bin/

# Add to PATH if needed
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Verify installation
lai --help
```

**Requirements**: Go 1.21+, OpenAI API key, Telegram bot token (optional), Email configuration (optional)

## 🚀 Quick Start

### Recommended: Unified Interface

The new `monitor` command provides a unified interface for all monitoring needs:

```bash
# Monitor a log file
lai monitor file /path/to/app.log

# Monitor Docker container logs
lai monitor command "docker logs webapp -f"

# Monitor build process with custom working directory
lai monitor command "npm run build" -w /path/to/project

# Run as daemon with custom name
lai monitor file /var/log/nginx/error.log -d -n "nginx-monitor"
```

### Legacy Commands (Still Supported)

For backward compatibility, the original commands remain available:

```bash
# Traditional file monitoring
lai start /path/to/app.log

# Traditional command monitoring
lai exec "docker logs webapp -f" -d

# Build process monitoring
lai exec "npm run build" --final-summary
```

**Real Example**: Monitor your production API logs and get notified when errors spike:

```bash
lai monitor file /var/log/api/error.log -d -n "api-monitor"
```

## 🛠️ Setup

### Option 1: Interactive Configuration (Recommended)

Launch the user-friendly TUI interface for guided configuration:

```bash
lai config interactive
```

The interactive interface will guide you through:

- Setting up your OpenAI API key
- Configuring notification providers
- Customizing monitoring preferences
- Testing your configuration

### Option 2: Command Line Configuration

1. **Configure OpenAI**:

   ```bash
   lai config set notifications.openai.api_key "sk-your-key"
   ```

2. **Configure Notifications** (choose from multiple providers):

   **Telegram Notifications**:

   ```bash
   lai config set notifications.providers.telegram.bot_token "123456:ABC-DEF"
   lai config set notifications.providers.telegram.chat_id "-100123456789"
   ```

   **Email Notifications**:

   ```bash
   lai config set notifications.providers.email.smtp_host "smtp.gmail.com"
   lai config set notifications.providers.email.smtp_port "587"
   lai config set notifications.providers.email.username "your-email@gmail.com"
   lai config set notifications.providers.email.password "your-app-password"
   lai config set notifications.providers.email.from_email "your-email@gmail.com"
   lai config set notifications.providers.email.to_emails '["recipient1@gmail.com", "recipient2@gmail.com"]'
   lai config set notifications.providers.email.subject "Lai Log Alert"
   ```

   **Discord Notifications**:

   ```bash
   lai config set notifications.providers.discord.webhook_url "https://discord.com/api/webhooks/..."
   ```

   **Slack Notifications**:

   ```bash
   lai config set notifications.providers.slack.webhook_url "https://hooks.slack.com/services/..."
   ```

3. **Configure AI response language** (optional):

   ```bash
   lai config set defaults.language "Chinese"  # Chinese
   lai config set defaults.language "Spanish"  # Spanish  
   lai config set defaults.language "English"  # English (default)
   ```

4. **Start monitoring**:

   ```bash
   # Use all configured notifiers
   lai monitor file /path/to/your.log
   
   # Use specific notifiers only
   lai monitor file /path/to/your.log --notifiers telegram,email
   
   # Use only email notifications
   lai monitor file /path/to/your.log --notifiers email
   
   # Monitor command output
   lai monitor command "docker logs myapp -f" --notifiers discord
   ```

## 📖 Common Use Cases

### Monitor Application Logs

```bash
# Background monitoring with custom name
lai monitor file /var/log/nginx/error.log -d -n "nginx-errors"

# View what's being monitored
lai list

# Check monitoring logs
lai logs nginx-errors -f
```

### Monitor Docker Containers

```bash
# Monitor specific container
lai monitor command "docker logs webapp -f" -d -n "webapp-monitor"

# Monitor with custom thresholds
lai monitor command "docker logs db -f" --line-threshold 5 --interval 10s
```

### Monitor Build/CI Processes

```bash
# Get summary when build completes
lai monitor command "npm run build" --final-summary

# Monitor tests with error detection
lai monitor command "npm test" -l 3 -i 15s
```

## 🔧 Process Management

```bash
lai list           # Show all running monitors
lai stop <name>    # Stop a monitor
lai resume <name>  # Restart a stopped monitor
lai clean          # Remove stopped entries
```

## 📚 Documentation

- **[Development Guide](docs/DEVELOPMENT.md)** - Building, testing, and contributing
- **[Configuration Reference](docs/CONFIGURATION.md)** - Complete configuration options
- **[Architecture Overview](docs/ARCHITECTURE.md)** - Project structure and design

## 📋 Roadmap

### Recently Completed ✅

- [x] **Unified monitoring interface** - Single `monitor` command for all monitoring needs
- [x] **Interactive configuration TUI** - User-friendly configuration interface
- [x] **Discord and Slack notifications** - Expanded notification provider support
- [x] **Unified notification system** - Improved provider configuration system
- [x] **Cross-platform improvements** - Enhanced Windows compatibility
- [x] **Configuration metadata system** - Better configuration validation and documentation

### Upcoming Features 🚀

- [ ] Webhook notifications support
- [ ] Advanced log filtering and pattern matching
- [ ] Integration with popular monitoring tools (Prometheus, Grafana)

### Legacy Features ✅

- [x] **Email notification support** with SMTP configuration
- [x] **Multi-notifier support** - enable/disable via config or command line
- [x] **Customizable message templates** for each notification channel
- [x] Multi-language support for AI responses
- [x] Error-only notification mode (filter out normal logs, notify only on errors/exceptions)

## 🤝 Contributing

> Note: Usage on Windows platforms are experimental.

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality  
4. Ensure all tests pass: `make test`
5. Submit a pull request

## 📄 License

See LICENSE file for details.
