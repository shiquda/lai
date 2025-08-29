## Lai: AI Powered Log Monitoring and Notification Tool

Lai is an AI-powered log monitoring tool that watches log files, generates intelligent summaries using OpenAI API when new content is detected, and sends notifications via Telegram.

## Features

- **Intelligent Log Analysis**: Uses OpenAI GPT models to analyze and summarize log content
- **Real-time Monitoring**: Continuously monitors log files for changes
- **Telegram Integration**: Sends formatted notifications directly to Telegram
- **Configurable Thresholds**: Set custom line thresholds and check intervals
- **Global Configuration**: Manage settings globally or per-project

## Installation

```bash
# Build from source
go build -o lai

# Or run directly
go run main.go
```

## Usage

### Basic Usage

```bash
# Start monitoring a log file
./lai start /path/to/logfile.log

# With custom settings
./lai start /path/to/logfile.log --line-threshold 10 --interval 30s --chat-id "-100123456789"
```

### Configuration

Lai uses a global configuration file at `~/.lai/config.yaml`:

```yaml
notifications:
  openai:
    api_key: "your-openai-api-key"
    base_url: "https://api.openai.com/v1"
    model: "gpt-4o"
  telegram:
    bot_token: "your-telegram-bot-token"
    chat_id: "your-default-chat-id"

defaults:
  line_threshold: 10
  check_interval: "30s"
  chat_id: "your-default-chat-id"
```

### Configuration Management

```bash
# Set OpenAI API key
./lai config set notifications.openai.api_key "sk-your-key"

# Set Telegram bot token
./lai config set notifications.telegram.bot_token "123456:ABC-DEF"

# Set default chat ID
./lai config set defaults.chat_id "-100123456789"

# View current configuration
./lai config list

# Reset configuration to defaults
./lai config reset
```

## Development

### Testing

The project includes comprehensive tests covering all major components:

#### Quick Test Run
```bash
# Run all tests
./scripts/test-quick.sh

# Or manually
go test ./...
```

#### Comprehensive Test Run
```bash
# Run full test suite with coverage
./scripts/test.sh
```

#### Test Structure
- **Unit Tests**: Located alongside source files (`*_test.go`)
- **Integration Tests**: End-to-end workflow testing (`integration_test.go`)
- **Test Coverage**: Aim for ≥80% coverage across all packages

#### Individual Package Testing
```bash
# Test specific packages
go test ./internal/collector -v
go test ./internal/config -v
go test ./internal/notifier -v
go test ./internal/summarizer -v
```

### Project Structure

```
lai/
├── main.go                          # Application entry point
├── cmd/                            # CLI commands
├── internal/                       # Internal packages
│   ├── collector/                  # Log file monitoring
│   ├── config/                     # Configuration management
│   ├── notifier/                   # Telegram notifications
│   ├── summarizer/                 # OpenAI integration
│   └── testutils/                  # Shared testing utilities
├── testdata/                       # Test data files
├── scripts/                        # Build and test scripts
└── integration_test.go             # Integration tests
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass: `./scripts/test.sh`
5. Submit a pull request

### Development Commands

```bash
# Download dependencies
go mod download

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Format code
gofmt -w .

# Run static analysis
go vet ./...
```

## Requirements

- Go 1.21.5 or later
- OpenAI API access
- Telegram bot token

## License

See LICENSE file for details.

---

**Status**: Under active development
