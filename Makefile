# Berga CLI Makefile

# Variables
BINARY_NAME=berga
MAIN_PACKAGE=.
BUILD_DIR=dist
VERSION?=1.0.0

# Go build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"
GCFLAGS=-gcflags=all="-l -B"

# Default target
.PHONY: all
all: clean build

# Build for current platform
.PHONY: build
build:
	@echo "Building berga for current platform..."
	go build $(LDFLAGS) $(GCFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Build for all platforms
.PHONY: build-all
build-all: clean
	@echo "Building berga for all platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) $(GCFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	GOOS=windows GOARCH=386 go build $(LDFLAGS) $(GCFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-386.exe $(MAIN_PACKAGE)
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) $(GCFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) $(GCFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) $(GCFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=386 go build $(LDFLAGS) $(GCFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-386 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) $(GCFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)

# Install to local system
.PHONY: install
install: build
	@echo "Installing berga..."
ifeq ($(OS),Windows_NT)
	@echo "To install globally on Windows:"
	@echo "1. Copy $(BUILD_DIR)/$(BINARY_NAME).exe to a directory in your PATH"
	@echo "2. Or run: copy $(BUILD_DIR)\\$(BINARY_NAME).exe C:\\Windows\\System32\\"
else
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "berga installed to /usr/local/bin/$(BINARY_NAME)"
endif

# Development install (symlink for easy updates during development)
.PHONY: dev-install
dev-install: build
	@echo "Installing berga for development..."
ifeq ($(OS),Windows_NT)
	@echo "Development install not supported on Windows. Use 'make install' instead."
else
	ln -sf $(PWD)/$(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "berga dev-installed (symlinked) to /usr/local/bin/$(BINARY_NAME)"
endif

# Test
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Lint
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Format
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)

# Run
.PHONY: run
run:
	@echo "Running berga..."
	go run $(MAIN_PACKAGE)

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build       - Build for current platform"
	@echo "  build-all   - Build for all platforms"
	@echo "  install     - Install to system"
	@echo "  dev-install - Install for development (symlink)"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linter"
	@echo "  fmt         - Format code"
	@echo "  clean       - Clean build artifacts"
	@echo "  run         - Run without building"
	@echo "  help        - Show this help"

# Ensure directories exist
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)
