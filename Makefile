# Makefile for cclogviewer

# Binary name
BINARY_NAME=cclogviewer

# Build directory
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build variables
VERSION?=1.0.0
BUILD_TIME=$(shell date +%FT%T%z)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Installation directory
PREFIX?=/usr/local
INSTALL_DIR=$(PREFIX)/bin

# Default target
.DEFAULT_GOAL := build

# Create build directory
$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

# Build the binary
build: $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/cclogviewer

# Build with version info
build-release: $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/cclogviewer

# Run the application with example
run: build
	./$(BUILD_DIR)/$(BINARY_NAME) -input example.jsonl

# Run quick view (auto-open)
run-quick: build
	./$(BUILD_DIR)/$(BINARY_NAME) -input example.jsonl

# Run with specific output
run-output: build
	./$(BUILD_DIR)/$(BINARY_NAME) -input example.jsonl -output test_output.html -open

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)"
	@mkdir -p $(INSTALL_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/
	@chmod 755 $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installation complete. You can now run '$(BINARY_NAME)' from anywhere."

# Uninstall the binary
uninstall:
	@echo "Removing $(BINARY_NAME) from $(INSTALL_DIR)"
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Uninstall complete."

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f test_output.html
	rm -f example_*.html

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	$(GOCMD) fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux: $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 -v ./cmd/cclogviewer
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 -v ./cmd/cclogviewer

build-darwin: $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 -v ./cmd/cclogviewer
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 -v ./cmd/cclogviewer

build-windows: $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe -v ./cmd/cclogviewer

# Create release archives
release: build-all
	mkdir -p dist
	tar -czf dist/$(BINARY_NAME)-linux-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-amd64 -C .. README.md
	tar -czf dist/$(BINARY_NAME)-linux-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-linux-arm64 -C .. README.md
	tar -czf dist/$(BINARY_NAME)-darwin-amd64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-amd64 -C .. README.md
	tar -czf dist/$(BINARY_NAME)-darwin-arm64.tar.gz -C $(BUILD_DIR) $(BINARY_NAME)-darwin-arm64 -C .. README.md
	cd $(BUILD_DIR) && zip ../dist/$(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe && cd .. && zip -j dist/$(BINARY_NAME)-windows-amd64.zip README.md

# Show help
help:
	@echo "Available targets:"
	@echo "  make build          - Build the binary"
	@echo "  make run            - Build and run with example file"
	@echo "  make run-quick      - Build and run with auto-open"
	@echo "  make run-output     - Build and run with specific output"
	@echo "  make install        - Install binary to $(INSTALL_DIR)"
	@echo "  make uninstall      - Remove binary from $(INSTALL_DIR)"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make deps           - Download and tidy dependencies"
	@echo "  make fmt            - Format Go code"
	@echo "  make lint           - Run linter (requires golangci-lint)"
	@echo "  make build-all      - Build for all platforms"
	@echo "  make release        - Create release archives"
	@echo ""
	@echo "Installation prefix can be changed with PREFIX:"
	@echo "  make install PREFIX=/opt/local"

.PHONY: build build-release run run-quick run-output install uninstall clean deps fmt lint build-all build-linux build-darwin build-windows release help