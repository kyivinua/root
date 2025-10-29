.PHONY: help build install test test-coverage test-race lint fmt clean run release

# Variables
BINARY_NAME=docgen
VERSION=1.0.0
BUILD_DIR=./bin
GO_FILES=$(shell find . -name '*.go' -type f -not -path './vendor/*')
MAIN_PATH=./cmd/docgen

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m

help: ## Show this help message
	@echo '$(COLOR_BOLD)Available targets:$(COLOR_RESET)'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'

build: ## Build the binary
	@echo '$(COLOR_BOLD)Building $(BINARY_NAME)...$(COLOR_RESET)'
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo '$(COLOR_GREEN)✓ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)'

install: build ## Install the binary to system
	@echo '$(COLOR_BOLD)Installing $(BINARY_NAME)...$(COLOR_RESET)'
	@go install -ldflags="-X main.version=$(VERSION)" $(MAIN_PATH)
	@echo '$(COLOR_GREEN)✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)$(COLOR_RESET)'

test: ## Run tests
	@echo '$(COLOR_BOLD)Running tests...$(COLOR_RESET)'
	@go test -v ./...
	@echo '$(COLOR_GREEN)✓ Tests complete$(COLOR_RESET)'

test-coverage: ## Run tests with coverage
	@echo '$(COLOR_BOLD)Running tests with coverage...$(COLOR_RESET)'
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@echo '$(COLOR_GREEN)✓ Coverage report: coverage.out$(COLOR_RESET)'

coverage-html: test-coverage ## Generate HTML coverage report
	@echo '$(COLOR_BOLD)Generating HTML coverage report...$(COLOR_RESET)'
	@go tool cover -html=coverage.out -o coverage.html
	@echo '$(COLOR_GREEN)✓ HTML coverage report: coverage.html$(COLOR_RESET)'

test-race: ## Run tests with race detector
	@echo '$(COLOR_BOLD)Running tests with race detector...$(COLOR_RESET)'
	@go test -race -v ./...
	@echo '$(COLOR_GREEN)✓ Race detector tests complete$(COLOR_RESET)'

lint: ## Run linter (if golangci-lint is installed)
	@echo '$(COLOR_BOLD)Running linter...$(COLOR_RESET)'
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo '$(COLOR_GREEN)✓ Linting complete$(COLOR_RESET)'; \
	else \
		echo '$(COLOR_YELLOW)⚠ golangci-lint not installed, skipping$(COLOR_RESET)'; \
	fi

fmt: ## Format Go code
	@echo '$(COLOR_BOLD)Formatting code...$(COLOR_RESET)'
	@gofmt -w $(GO_FILES)
	@echo '$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)'

clean: ## Clean build artifacts
	@echo '$(COLOR_BOLD)Cleaning...$(COLOR_RESET)'
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo '$(COLOR_GREEN)✓ Clean complete$(COLOR_RESET)'

run: build ## Build and run the application
	@$(BUILD_DIR)/$(BINARY_NAME)

release: clean ## Create release builds for multiple platforms
	@echo '$(COLOR_BOLD)Creating release builds...$(COLOR_RESET)'
	@mkdir -p $(BUILD_DIR)/release
	@GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=linux GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo '$(COLOR_GREEN)✓ Release builds complete: $(BUILD_DIR)/release/$(COLOR_RESET)'

mod-tidy: ## Tidy go modules
	@echo '$(COLOR_BOLD)Tidying modules...$(COLOR_RESET)'
	@go mod tidy
	@echo '$(COLOR_GREEN)✓ Modules tidied$(COLOR_RESET)'

mod-download: ## Download go modules
	@echo '$(COLOR_BOLD)Downloading modules...$(COLOR_RESET)'
	@go mod download
	@echo '$(COLOR_GREEN)✓ Modules downloaded$(COLOR_RESET)'

.DEFAULT_GOAL := help
