.PHONY: build clean install test lint run help

# Variables
BINARY_NAME=resumectl
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build directories
BUILD_DIR=bin
CMD_DIR=./cmd/resumectl

## help: Show this help message
help:
	@echo "resumectl - Makefile Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(GOPATH)/bin/$(BINARY_NAME) $(CMD_DIR)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf output

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## lint: Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## run: Build and run with default example
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) generate --html

## generate: Generate resume (HTML + PDF)
generate: build
	./$(BUILD_DIR)/$(BINARY_NAME) generate

## generate-html: Generate resume (HTML only)
generate-html: build
	./$(BUILD_DIR)/$(BINARY_NAME) generate --html

## generate-pdf: Generate resume (PDF only)
generate-pdf: build
	./$(BUILD_DIR)/$(BINARY_NAME) generate --pdf

## show: Show resume in terminal
show: build
	./$(BUILD_DIR)/$(BINARY_NAME) show

## validate: Validate resume YAML file
validate: build
	./$(BUILD_DIR)/$(BINARY_NAME) validate

## themes: List available themes
themes: build
	./$(BUILD_DIR)/$(BINARY_NAME) themes

## version: Show version
version: build
	./$(BUILD_DIR)/$(BINARY_NAME) version

## release: Build for multiple platforms
release: clean
	@echo "Building releases..."
	@mkdir -p $(BUILD_DIR)/releases
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/releases/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	@echo "Releases built in $(BUILD_DIR)/releases/"
