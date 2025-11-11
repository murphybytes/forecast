.PHONY: build test clean run help coverage

# Binary name
BINARY_NAME=forecast

# Build the forecast server
build:
	@echo "Building forecast server..."
	go build -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

# Test the forecast server
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -cover ./...
	@echo ""
	@echo "Generating detailed coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# Run the forecast server
run: build
	@echo "Starting forecast server..."
	./$(BINARY_NAME)

# Install dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted"

# Lint code (requires golangci-lint to be installed)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: brew install golangci-lint"; \
	fi

# Build and test
all: clean build test

# Display help
help:
	@echo "Available targets:"
	@echo "  build     - Build the forecast server binary"
	@echo "  test      - Run unit tests"
	@echo "  coverage  - Run tests with coverage report"
	@echo "  clean     - Remove build artifacts"
	@echo "  run       - Build and run the forecast server"
	@echo "  deps      - Download and tidy dependencies"
	@echo "  fmt       - Format Go code"
	@echo "  lint      - Lint Go code (requires golangci-lint)"
	@echo "  all       - Clean, build, and test"
	@echo "  help      - Display this help message"
