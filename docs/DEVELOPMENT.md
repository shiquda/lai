# Development Guide

This guide covers building, testing, and contributing to Lai.

## Quick Start

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

## Building

### Using Makefile (Recommended)

```bash
# Build the application
make build

# Install to GOPATH/bin
make install

# Clean build artifacts and test cache
make clean
```

### Using Go Directly

```bash
# Build with Go
go build -o lai

# Run directly without building
go run main.go
```

### Dependencies

```bash
# Download dependencies (using Makefile)
make deps

# Or manually with Go
go mod download

# Update dependencies
go mod tidy

# Vendor dependencies (if needed)
go mod vendor
```

## Testing

The project includes comprehensive tests covering all major components with 46.4% overall test coverage.

### Quick Test Run

```bash
# Run tests quickly (no coverage) - recommended for development
make test-quick

# Or manually
go test ./... -v
```

### Comprehensive Test Run

```bash
# Run full test suite with coverage and quality checks
make test

# This runs the test script at ./scripts/test-simple.sh
```

### Coverage Reports

```bash
# Generate HTML coverage report
make test-coverage

# Opens coverage.html - view in browser
# Current project coverage: 46.4%
```

### Individual Package Testing

```bash
# Test specific packages with verbose output
go test ./internal/collector -v
go test ./internal/config -v  
go test ./internal/daemon -v
go test ./internal/notifier -v
go test ./internal/summarizer -v
```

### Test Structure

- **Unit Tests**: Located alongside source files (`*_test.go`)
- **Integration Tests**: End-to-end workflow testing (`integration_test.go`)
- **Test Coverage**: Current coverage is 46.4% across all packages
- **Daemon Management Tests**: Complete test suites for all 5 daemon commands

## Code Quality

```bash
# Format all Go code
make fmt

# Run static analysis
make vet

# Full quality check pipeline
make clean fmt vet test
```

## Development Workflow

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
lai --help

# 6. Clean up when done
make clean
```

## Requirements

- Go 1.21.5 or later
- OpenAI API access
- Telegram bot token

## Contributing

### Getting Started

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Write tests for new functionality
5. Ensure all tests pass: `make test`
6. Run code quality checks: `make fmt vet`
7. Submit a pull request

### Adding New Components

When adding new functionality, follow the existing package structure:

- Place domain-specific logic in `internal/`
- Use dependency injection patterns seen in `cmd/start.go`
- Implement proper error handling with wrapped errors using `fmt.Errorf`

### Configuration Changes

- Update the `Config` struct in `internal/config/config.go`
- Add validation in the `Validate()` method
- Update `config.example.yaml` with new fields

### Adding New Notification Channels

Follow the pattern established in `internal/notifier/`:

- Create a new struct with the required credentials
- Implement methods for sending messages
- Integrate into the main flow in `cmd/start.go`

## Debug Guidelines

- Debug logs and comments should be in English
- Use structured logging where possible
- Include context in error messages

## Important Notes

- Follow the existing code style and patterns
- Add tests for any new functionality
- Update documentation when making changes