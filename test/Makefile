# Test Makefile for AI Gateway Hub

.PHONY: all unit integration e2e coverage clean help

# Default target
all: unit integration e2e

# Run unit tests
unit:
	@echo "Running unit tests..."
	@cd .. && ./scripts/go-test.sh unit

# Run integration tests
integration:
	@echo "Running integration tests..."
	@cd .. && ./scripts/go-test.sh integration

# Run E2E tests
e2e:
	@echo "Running E2E tests..."
	@cd .. && ./scripts/go-test.sh e2e

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@cd .. && ./scripts/go-test.sh coverage

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	@cd .. && ./scripts/go-test.sh clean

# Show help
help:
	@echo "Available targets:"
	@echo "  unit        - Run unit tests"
	@echo "  integration - Run integration tests"
	@echo "  e2e         - Run E2E tests"
	@echo "  coverage    - Run tests with coverage report"
	@echo "  all         - Run all tests (default)"
	@echo "  clean       - Clean test artifacts"
	@echo "  help        - Show this help message"