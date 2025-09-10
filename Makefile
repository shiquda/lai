# Makefile for Lai project

.PHONY: help build build-all install test test-quick test-coverage clean fmt vet deps act-test act-test-build

# Detect OS for cross-platform compatibility
ifeq ($(OS),Windows_NT)
    # Windows
    EXECUTABLE_EXT = .exe
    DATE_CMD = powershell -Command "Get-Date -UFormat '+%%Y-%%m-%%d_%%H:%%M:%%S'" 2>nul || echo unknown
    NULL_REDIRECT = 2>nul
else
    # Unix-like (Linux, macOS, etc.)
    EXECUTABLE_EXT = 
    DATE_CMD = date -u '+%Y-%m-%d_%H:%M:%S'
    NULL_REDIRECT = 2>/dev/null
endif

# Version information - cross-platform
VERSION ?= $(shell git describe --tags --always --dirty $(NULL_REDIRECT) || echo dev)
BUILD_TIME ?= $(shell $(DATE_CMD))
GIT_COMMIT ?= $(shell git rev-parse --short HEAD $(NULL_REDIRECT) || echo unknown)

# Build flags - cross-platform
LDFLAGS = -ldflags "-X github.com/shiquda/lai/internal/version.Version=$(VERSION) -X github.com/shiquda/lai/internal/version.BuildTime=$(BUILD_TIME) -X github.com/shiquda/lai/internal/version.GitCommit=$(GIT_COMMIT)"

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build the application for current platform"
	@echo "  build-all    - Build for all platforms (Linux, Windows, macOS)"
	@echo "  install      - Install the application to GOPATH/bin"
	@echo "  test         - Run all tests with coverage and quality checks"
	@echo "  test-quick   - Run tests quickly (no coverage)"
	@echo "  test-coverage- Run tests with coverage report only"
	@echo "  clean        - Clean build artifacts and test cache"
	@echo "  fmt          - Format all Go code"
	@echo "  vet          - Run go vet static analysis"
	@echo "  deps         - Download and verify dependencies"
	@echo "  act-test     - Test GitHub Actions locally with act (full workflow)"
	@echo "  act-test-build - Test cross-platform build job with act"

# Build the application for current platform
build:
	@echo "🔨 Building lai..."
	@go build $(LDFLAGS) -o lai$(EXECUTABLE_EXT)
	@echo "✅ Built successfully."

# Build for all platforms
build-all:
	@echo "🔨 Building lai for all platforms..."
	@echo "Building Linux AMD64..."
ifeq ($(OS),Windows_NT)
	@set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build $(LDFLAGS) -o lai-$(VERSION)-linux-amd64
	@echo "Building Windows AMD64..."
	@set GOOS=windows&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build $(LDFLAGS) -o lai-$(VERSION)-windows-amd64.exe
	@echo "Building macOS AMD64..."
	@set GOOS=darwin&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build $(LDFLAGS) -o lai-$(VERSION)-darwin-amd64
	@echo "Building macOS ARM64..."
	@set GOOS=darwin&& set GOARCH=arm64&& set CGO_ENABLED=0&& go build $(LDFLAGS) -o lai-$(VERSION)-darwin-arm64
else
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o lai-$(VERSION)-linux-amd64
	@echo "Building Windows AMD64..."
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o lai-$(VERSION)-windows-amd64.exe  
	@echo "Building macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o lai-$(VERSION)-darwin-amd64
	@echo "Building macOS ARM64..."
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o lai-$(VERSION)-darwin-arm64
endif
	@echo "✅ Built all platforms successfully:"
	@echo "  - lai-$(VERSION)-linux-amd64"
	@echo "  - lai-$(VERSION)-windows-amd64.exe"
	@echo "  - lai-$(VERSION)-darwin-amd64" 
	@echo "  - lai-$(VERSION)-darwin-arm64"

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
ifeq ($(OS),Windows_NT)
	@-del lai.exe 2>nul || echo ""
	@-del coverage.out 2>nul || echo ""
	@-del coverage.html 2>nul || echo ""
else
	@rm -f lai coverage.out coverage.html
endif

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

# Test GitHub Actions with act
act-test:
	@echo "🎭 Testing GitHub Actions with act..."
	@echo "Note: This will test the full workflow with mock release tag"
	@act push -e .github/act_push_event.json --artifact-server-path /tmp/artifacts

# Test only build job with act (now simplified single job)
act-test-build:
	@echo "🎭 Testing build job with act..."
	@echo "Note: This tests the cross-platform build job"
	@act push -j build -e .github/act_push_event.json --artifact-server-path /tmp/artifacts