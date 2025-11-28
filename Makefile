.PHONY: build clean test run install help

# Binary name
BINARY_NAME=wanikani-api
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/wanikani-api

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/wanikani-api
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/wanikani-api
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/wanikani-api
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/wanikani-api

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f *.db *.db-shm *.db-wal

# Run tests
test:
	@echo "Running unit tests..."
	$(GOTEST) -v ./...

# Run unit tests except property tests
test-short:
	@echo "Running unit tests wo property tests..."
	$(GOTEST) -short -v ./...

# Run integration tests (requires .env with WANIKANI_API_TOKEN)
test-integration: 
	@echo "Running integration tests..."
	@echo "Note: Requires WANIKANI_API_TOKEN in .env or environment"
	$(GOTEST) -tags=integration -v ./...

# Run all tests (unit + integration)
test-all:
	@echo "Running all tests (unit + integration)..."
	$(GOTEST) -v ./...
	$(GOTEST) -tags=integration -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run the application
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install dependencies
install:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

# Display help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application binary to bin/"
	@echo "  build-all     - Build for multiple platforms (Linux, macOS, Windows)"
	@echo "  clean         - Remove build artifacts and test databases"
	@echo "  test          - Run unit tests"
	@echo "  test-short    - Run unit tests wo property tests"
	@echo "  test-integration - Run integration tests (requires .env with API token)"
	@echo "  test-all      - Run both unit and integration tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  run           - Build and run the application"
	@echo "  install       - Install/update dependencies"
	@echo "  fmt           - Format Go code"
	@echo "  lint          - Run linter (requires golangci-lint)"
	@echo "  help          - Display this help message"
