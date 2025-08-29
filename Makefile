# Makefile for Lai project

.PHONY: help build install test test-quick test-coverage clean fmt vet deps

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
	@go build -o lai
	@echo "✅ Built successfully."

# Install the application
install:
	@echo "📦 Installing lai..."
	@go install .
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

# Static analysis
vet:
	@echo "🔍 Running static analysis..."
	@go vet ./...

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	@go mod download
	@go mod verify