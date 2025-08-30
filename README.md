# Lai: AI-Powered Log Monitoring

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/doc/install)
[![AI Powered](https://img.shields.io/badge/AI-Powered-brightgreen.svg)]()
[![Telegram](https://img.shields.io/badge/Notifications-Telegram-blue.svg)](https://telegram.org/)
[![License](https://img.shields.io/badge/License-AGPL--3.0-yellow.svg)](LICENSE)

Stop manually checking logs. Let AI watch, analyze, and notify you when something important happens.

> Note: This project is under active development. Any kind of contributions are welcome!

## ✨ Core Features

- **🤖 AI-Powered Analysis**: GPT automatically summarizes log changes and identifies critical issues
- **📱 Instant Notifications**: Get smart alerts via Telegram when errors or important events occur
- **🔄 Universal Monitoring**: Watch any log file or command output (Docker logs, application output, build processes)
- **🔌 Hot-Pluggable**: No code changes required - works with any existing project or application

## ⚡ Installation

### Option 1: Install from Release (Recommended)

```bash
# Download latest release
wget https://github.com/shiquda/lai/releases/latest/download/lai-linux-amd64.tar.gz

# Extract archive
tar -xzf lai-linux-amd64.tar.gz

# Create local bin directory and move binary (rename to lai)
mkdir -p ~/.local/bin
mv lai-linux-amd64 ~/.local/bin/lai

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

**Requirements**: Go 1.21+, OpenAI API key, Telegram bot token

## 🚀 Quick Start

# Monitor a log file with instant AI summaries

```
lai start /path/to/app.log
```

# Monitor Docker container logs

```
lai exec "docker logs webapp -f" -d
```

# Monitor build process with completion summary

```
lai exec "npm run build" --final-summary
```

**Real Example**: Monitor your production API logs and get notified when errors spike:

```bash
lai start /var/log/api/error.log -d -n "api-monitor"
```

## 🛠️ Setup

1. **Configure OpenAI**:

   ```bash
   lai config set notifications.openai.api_key "sk-your-key"
   ```

2. **Configure Telegram**:

   ```bash
   lai config set notifications.telegram.bot_token "123456:ABC-DEF"
   lai config set defaults.chat_id "-100123456789"
   ```

3. **Start monitoring**:

   ```bash
   lai start /path/to/your.log
   ```

## 📖 Common Use Cases

### Monitor Application Logs

```bash
# Background monitoring with custom name
lai start /var/log/nginx/error.log -d -n "nginx-errors"

# View what's being monitored
lai list

# Check monitoring logs
lai logs nginx-errors -f
```

### Monitor Docker Containers

```bash
# Monitor specific container
lai exec "docker logs webapp -f" -d -n "webapp-monitor"

# Monitor with custom thresholds
lai exec "docker logs db -f" --line-threshold 5 --interval 10s
```

### Monitor Build/CI Processes

```bash
# Get summary when build completes
lai exec "npm run build" --final-summary

# Monitor tests with error detection
lai exec "npm test" -l 3 -i 15s
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

- [ ] Add more notification methods (email, Slack, Discord...)
- [ ] Support more customized settings (notification format, prompts, languages)
- [ ] Error-only notification mode (filter out normal logs, notify only on errors/exceptions)

## 🤝 Contributing

> Note: Usage on Windows platforms are experimental.

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality  
4. Ensure all tests pass: `make test`
5. Submit a pull request

## 📄 License

See LICENSE file for details.
