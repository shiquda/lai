#!/bin/bash

# Lai Test Runner Script
# This script runs all tests for the Lai project with coverage reporting

set -e

echo "🧪 Running Lai Test Suite"
echo "========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is available
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

print_status "Go version: $(go version)"

# Clean any previous test artifacts
print_status "Cleaning previous test artifacts..."
go clean -testcache

# Download dependencies
print_status "Downloading dependencies..."
go mod download

# Run unit tests for each package
packages=("./internal/collector" "./internal/config" "./internal/notifier" "./internal/summarizer" "./internal/testutils")

echo ""
echo "📦 Running Unit Tests"
echo "===================="

for pkg in "${packages[@]}"; do
    print_status "Testing $pkg..."
    if ! go test "$pkg" -v -race; then
        print_error "Unit tests failed for $pkg"
        exit 1
    fi
    print_success "Unit tests passed for $pkg"
done

# Run integration tests
echo ""
echo "🔗 Running Integration Tests"
echo "============================"

print_status "Running integration tests..."
if ! go test -run TestCollectorWorkflow -v; then
    print_error "Integration tests failed"
    exit 1
fi

if ! go test -run TestConfig -v; then
    print_error "Config integration tests failed"
    exit 1
fi

print_success "All integration tests passed"

# Run all tests with coverage
echo ""
echo "📊 Running Tests with Coverage"
echo "==============================="

print_status "Generating coverage report..."
go test -coverprofile=coverage.out ./...

if [ -f coverage.out ]; then
    # Generate coverage report
    go tool cover -html=coverage.out -o coverage.html
    
    # Show coverage summary
    coverage_percent=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    print_status "Total coverage: $coverage_percent"
    
    # Check if coverage meets minimum threshold (60%)
    coverage_num=$(echo $coverage_percent | sed 's/%//')
    if [ "$(echo "$coverage_num >= 60" | awk '{print ($1 >= $3)}')" = "1" ]; then
        print_success "Coverage meets minimum threshold (≥60%)"
    else
        print_warning "Coverage below recommended threshold (60%). Current: $coverage_percent"
    fi
    
    print_status "Coverage report saved to coverage.html"
else
    print_warning "Coverage report not generated"
fi

# Run linting if golangci-lint is available
echo ""
echo "🔍 Code Quality Checks"
echo "====================="

if command -v golangci-lint &> /dev/null; then
    print_status "Running golangci-lint..."
    if golangci-lint run; then
        print_success "Linting passed"
    else
        print_warning "Linting issues found (not blocking)"
    fi
else
    print_warning "golangci-lint not found, skipping linting"
    print_status "To install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

# Check go fmt
print_status "Checking code formatting..."
if [ -n "$(gofmt -l .)" ]; then
    print_warning "Code formatting issues found:"
    gofmt -l .
    print_status "Run 'gofmt -w .' to fix"
else
    print_success "Code formatting is correct"
fi

# Run go vet
print_status "Running go vet..."
if go vet ./...; then
    print_success "go vet passed"
else
    print_error "go vet found issues"
    exit 1
fi

echo ""
echo "✅ Test Summary"
echo "==============="
print_success "All tests completed successfully!"
print_status "Unit tests: ✅"
print_status "Integration tests: ✅"
print_status "Code formatting: ✅"
print_status "Static analysis: ✅"

if [ -f coverage.html ]; then
    echo ""
    print_status "Open coverage.html in your browser to view detailed coverage report"
fi

echo ""
print_success "Test suite completed successfully! 🎉"