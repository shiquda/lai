## Lai: AI Powered Log Monitoring and Notification Tool

Lai is an AI-powered log monitoring tool that watches log files, generates intelligent summaries using OpenAI API when new content is detected, and sends notifications via Telegram.

## Features

- **Intelligent Log Analysis**: Uses OpenAI GPT models to analyze and summarize log content
- **Real-time Monitoring**: Continuously monitors log files for changes
- **Command Output Monitoring**: Monitor any command's stdout/stderr in real-time (Docker logs, application output, etc.)
- **Program Exit Handling**: Configurable final summary on program termination
- **Telegram Integration**: Sends formatted notifications directly to Telegram
- **Daemon Process Management**: Run monitoring as background processes with full lifecycle management
- **Configurable Thresholds**: Set custom line thresholds and check intervals
- **Global Configuration**: Manage settings globally or per-project
- **Cross-Platform Support**: Native support for Windows, Linux, and macOS

## Platform Support

Lai runs natively on:
- **Windows** (amd64, arm64) - Full daemon process support
- **Linux** (amd64, arm64, 386) - Full Unix process management
- **macOS** (amd64, arm64) - Full Unix process management

### Platform-Specific Notes

#### Windows
- Daemon processes run as detached background processes
- Configuration files stored in `%USERPROFILE%\.lai\`
- Log paths use Windows format: `C:\path\to\logs\app.log`
- Process termination uses Windows-specific APIs

#### Unix/Linux/macOS
- Daemon processes use Unix fork/exec with session management
- Configuration files stored in `~/.lai/`
- Log paths use Unix format: `/path/to/logs/app.log`
- Process management uses Unix signals (SIGTERM, SIGKILL)

## Installation

```bash
# Build from source
make build

# Install to GOPATH/bin
make install

# Or using Go directly
go build -o lai

# Or run directly
go run main.go
```

## Usage

### Basic Usage

#### File Monitoring

```bash
# Unix/Linux/macOS
./lai start /path/to/logfile.log

# Windows (Command Prompt)
.\lai.exe start "C:\logs\app.log"

# Windows (PowerShell) 
.\lai.exe start "C:\logs\app.log"

# With custom settings
./lai start /path/to/logfile.log --line-threshold 10 --interval 30s --chat-id "-100123456789"

# Start as daemon process
./lai start /path/to/logfile.log -d

# Start daemon with custom name
./lai start /path/to/logfile.log -d -n "webapp-logs"
```

> Example: Consider a project that stores its logs in a local file. You can leverage `lai start /path/to/logfile.log -d -n "your-project"` command to monitor these logs dynamically, without requiring any modifications to the existing codebase.

#### Command Output Monitoring

Monitor any command's stdout/stderr in real-time:

```bash
# Monitor Docker container logs
./lai exec "docker logs container_name -f" --line-threshold 5 --interval 10s

# Monitor application output
./lai exec "python app.py" -l 3 -i 30s

# Monitor with custom name in daemon mode
./lai exec "docker logs webapp_container -f" -d -n "webapp-monitor"

# Enable final summary on program exit (default enabled)
./lai exec "npm run build" --final-summary

# Disable final summary
./lai exec "short-lived-command" --no-final-summary

# Run in specific working directory
./lai exec "make test" --workdir /path/to/project
```

> Example: Suppose you are running a Python script and wish to receive notifications upon encountering errors or a concise summary upon completion. Simply execute:
>
> ```bash
> lai exec "python3 main.py" --final-summary
> ```
>
> This accomplishes the desired outcome seamlessly.

### Daemon Process Management

```bash
# List running daemon processes
./lai list

# View daemon logs
./lai logs webapp-logs

# Follow daemon logs in real-time
./lai logs webapp-logs -f

# Stop a daemon process
./lai stop webapp-logs

# Resume a stopped daemon
./lai resume webapp-logs

# Clean up stopped daemon entries
./lai clean
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
  final_summary: true  # Send final summary when programs exit
```

### Configuration Management

```bash
# Set OpenAI API key
./lai config set notifications.openai.api_key "sk-your-key"

# Set Telegram bot token
./lai config set notifications.telegram.bot_token "123456:ABC-DEF"

# Set default chat ID
./lai config set defaults.chat_id "-100123456789"

# Set final summary default behavior
./lai config set defaults.final_summary true

# View current configuration
./lai config list

# Reset configuration to defaults
./lai config reset
```

## Development

### Quick Start

```bash
# View all available commands
make help

# Build the application
make build

# Download dependencies
make deps

# Format code
make fmt

# Run static analysis
make vet
```

### Testing

The project includes comprehensive tests covering all major components with 46.4% overall test coverage:

#### Quick Test Run
```bash
# Run tests quickly (no coverage)
make test-quick

# Or manually
go test ./... -v
```

#### Comprehensive Test Run
```bash
# Run full test suite with coverage and quality checks
make test
```

#### Generate Coverage Report
```bash
# Generate HTML coverage report
make test-coverage
# View coverage.html in browser
```

#### Test Structure
- **Unit Tests**: Located alongside source files (`*_test.go`)
- **Integration Tests**: End-to-end workflow testing (`integration_test.go`)
- **Test Coverage**: Current coverage is 46.4% across all packages
- **Daemon Management Tests**: Complete test suites for all 5 daemon commands

#### Individual Package Testing
```bash
# Test specific packages
go test ./internal/collector -v
go test ./internal/config -v
go test ./internal/notifier -v
go test ./internal/summarizer -v
go test ./internal/daemon -v
```

### Code Quality

```bash
# Clean build artifacts and test cache
make clean

# Format all code
make fmt

# Run static analysis
make vet

# Full quality check pipeline
make clean fmt vet test
```

### Project Structure

```
lai/
├── main.go                          # Application entry point
├── Makefile                         # Development workflow commands
├── cmd/                            # CLI commands
│   ├── start.go                    # Main start command with daemon support
│   ├── exec.go                     # Execute and monitor commands
│   ├── clean.go                    # Clean stopped daemon processes
│   ├── list.go                     # List running daemon processes  
│   ├── logs.go                     # View daemon process logs
│   ├── resume.go                   # Resume stopped daemon processes
│   └── stop.go                     # Stop running daemon processes
├── internal/                       # Internal packages
│   ├── collector/                  # Log file and command output monitoring
│   ├── config/                     # Configuration management
│   ├── daemon/                     # Daemon process lifecycle management
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
4. Ensure all tests pass: `make test`
5. Run code quality checks: `make fmt vet`
6. Submit a pull request

### Development Workflow

```bash
# 1. Setup development environment
make deps

# 2. Make your changes and format code
make fmt

# 3. Run static analysis
make vet

# 4. Run tests
make test

# 5. Build and test locally
make build
./lai --help

# 6. Clean up when done
make clean
```

## Requirements

- Go 1.21.5 or later
- OpenAI API access
- Telegram bot token

## To Do

- [ ] Add more notification methods, e.g. email
- [ ] Support more customized settings, e.g. notification format, prompts, languages etc.


## License

See LICENSE file for details.

---

**Status**: Under active development
