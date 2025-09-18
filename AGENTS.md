# Guidelines for AI Coding Agents

This document provides project context and expectations for assistants contributing code to this repository.

## Project Overview

Lai is an AI-powered log monitoring and notification tool written in Go. It watches log files, produces summaries with the OpenAI API, and dispatches notifications across Telegram, Email, Discord, and Slack.

## Architecture

- `main.go`: Entry point delegating to the Cobra CLI commands in `cmd/`
- `cmd/`: CLI surface area
  - `start.go`: Main monitoring command with daemon support (`-d` flag)
  - `clean.go`: Remove stopped daemon processes from the registry
  - `list.go`: Show running daemon processes and their state
  - `logs.go`: Inspect and follow daemon process log files
  - `resume.go`: Restart stopped daemon processes
  - `stop.go`: Terminate running daemon processes
- `internal/`: Core business logic organized by domain
  - `collector/`: Log monitoring and change detection
  - `daemon/`: Daemon lifecycle management and registry persistence
  - `summarizer/`: OpenAI integration for log analysis
  - `notifier/`: Multi-channel notification orchestration
  - `config/`: Configuration management and validation

## Development Workflow

- Prefer Makefile targets for common tasks:
  - `make help`: Discover available commands
  - `make build`: Build the project
  - `make deps`: Sync and verify Go module dependencies
- Equivalent Go commands are acceptable when necessary (e.g., `go build -o lai`).

### Running the Application

```bash
./lai start /path/to/logfile.log            # Run with default config
./lai start /path/to/logfile.log -d         # Run as a daemon
./lai start /path/to/logfile.log -d -n NAME # Custom daemon name
```

### Daemon Management Commands

```bash
./lai list             # Show running daemons
./lai logs NAME        # Tail daemon logs
./lai stop NAME        # Stop a daemon
./lai resume NAME      # Resume a stopped daemon
./lai clean            # Remove stopped daemon entries
```

### Configuration Management

The global configuration lives at `~/.lai/config.yaml`.

```bash
./lai config set notifications.openai.api_key "sk-your-key"
./lai config set notifications.providers.telegram.bot_token "123456:ABC-DEF"
./lai config set defaults.chat_id "-100123456789"
./lai config list
./lai config reset
```

If configuration corruption occurs during testing, restore it with:

```bash
cp ~/.lai/config.yaml.bak ~/.lai/config.yaml
```

## Dependency Management

- `make deps`: Download dependencies
- `go mod download`: Alternative manual dependency download
- `go mod tidy`: Update modules
- `go mod vendor`: Vendor dependencies if required

## Testing and Quality Gates

Always ensure a clean test run before committing changes:

- `make test-quick` or `go test ./... -v`: Fast feedback loop
- `make test`: Full suite including coverage and quality checks (`./scripts/test-simple.sh`)
- `make test-coverage`: Generate HTML coverage report (`coverage.html`)
- `go test ./internal/<package> -v`: Targeted package tests

For code quality:

- `make fmt`: Run `gofmt` across the codebase
- `make vet`: Run `go vet`
- `make clean`: Remove build artifacts and caches
- `make clean fmt vet test`: Full pipeline

## Contribution Expectations

- Follow Go formatting conventions; run `gofmt` on modified files.
- Keep tests green. Run the relevant `make` or `go` commands listed above after changes.
- Update documentation (e.g., `README.md`, `docs/`) when user-facing behavior changes.
- Use clear, English-language debug logs and comments.

## Notes for AI Agents

- Do not create new git branches.
- Use conventional commits and keep the working tree clean before finishing.
- If you must amend the latest commit, use `git commit --amend --no-edit`.
- Obey user-provided instructions (including language preferences) alongside this document.
