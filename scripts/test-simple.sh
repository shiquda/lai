#!/bin/bash

# Simplified Test Runner for Lai Project
set -e

echo "🧪 Running Lai Test Suite"
echo "========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Check Go installation
if ! command -v go &> /dev/null; then
    print_error "Go is not installed"
    exit 1
fi

print_status "Go version: $(go version)"

# Clean test cache
print_status "Cleaning test cache..."
go clean -testcache

# Download dependencies
print_status "Downloading dependencies..."
go mod download

# Run all tests
echo ""
echo "🧪 Running All Tests"
echo "===================="

print_status "Running tests..."
if go test ./... -v; then
    print_success "All tests passed!"
else
    print_error "Some tests failed"
    exit 1
fi

# Generate coverage report
echo ""
echo "📊 Coverage Report"
echo "=================="

print_status "Generating coverage..."
go test -coverprofile=coverage.out ./... > /dev/null 2>&1

if [ -f coverage.out ]; then
    go tool cover -html=coverage.out -o coverage.html
    coverage_summary=$(go tool cover -func=coverage.out | grep total)
    print_status "Coverage summary: $coverage_summary"
    print_status "Detailed report saved to coverage.html"
else
    print_warning "Coverage report not generated"
fi

# Basic code quality checks
echo ""
echo "🔍 Code Quality"
echo "==============="

print_status "Running go vet..."
if go vet ./...; then
    print_success "Static analysis passed"
else
    print_warning "Static analysis found issues"
fi

print_status "Checking formatting..."
unformatted=$(gofmt -l .)
if [ -z "$unformatted" ]; then
    print_success "Code formatting is correct"
else
    print_warning "Code formatting issues found:"
    echo "$unformatted"
fi

echo ""
print_success "Test suite completed! ✅"