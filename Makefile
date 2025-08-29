# Makefile for Lai project

.PHONY: help build install test test-quick test-coverage clean fmt vet deps

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS = -ldflags "-X github.com/shiquda/lai/cmd.Version=$(VERSION) -X github.com/shiquda/lai/cmd.BuildTime=$(BUILD_TIME) -X github.com/shiquda/lai/cmd.GitCommit=$(GIT_COMMIT)"

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the application"
	@echo "  install      - Install the application to GOPATH/bin"
	@echo "  test         - Run all tests with coverage and quality checks"
	@echo "  test-quick   - Run tests quickly (no coverage)"
	@echo "  test-coverage- Run tests with coverage report only"
	@echo "  clean        - Clean build artifacts and test cache"
	@echo "  fmt          - Format all Go code"
	@echo "  vet          - Run go vet static analysis"
	@echo "  deps         - Download and verify dependencies"

# Build the application
build:
	@echo "🔨 Building lai..."
	@go build $(LDFLAGS) -o lai
	@echo "✅ Built successfully."

# Install the application
install:
	@echo "📦 Installing lai..."
	@go install $(LDFLAGS) .
	@echo "✅ Installed successfully to GOPATH/bin"

# Quick test run (for development)
test-quick:
	@echo "🚀 Running quick tests..."
	@go test ./... -v

# Full test suite with coverage and quality checks
test:
	@echo "🧪 Running comprehensive test suite..."
	@./scripts/test-simple.sh

# Coverage report only
test-coverage:
	@echo "📊 Generating test coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@go clean -testcache
	@rm -f lai coverage.out coverage.html

# Format code
fmt:
	@echo "📐 Formatting code..."
	@gofmt -w .
	@echo "✅ Formatted."

# Static analysis
vet:
	@echo "🔍 Running static analysis..."
	@go vet ./...

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod verify