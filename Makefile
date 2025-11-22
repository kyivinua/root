.PHONY: help build install test test-coverage test-race lint fmt clean run release proto-lint proto-breaking proto-build proto-docs proto-ci build-proto-docs build-runtime build-enricher run-runtime run-enricher test-protoctx test-pipeline test-enricher

# Variables
BINARY_NAME=docgen
VERSION=1.0.0
BUILD_DIR=./bin
GO_FILES=$(shell find . -name '*.go' -type f -not -path './vendor/*')
MAIN_PATH=./cmd/docgen
PROTO_DOCS_BIN=$(BUILD_DIR)/proto-docs
RUNTIME_BIN=$(BUILD_DIR)/runtime
ENRICHER_BIN=$(BUILD_DIR)/protodocs-enricher

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

# Proto documentation pipeline targets

proto-lint: ## Run buf lint on proto files
	@echo '$(COLOR_BOLD)Running buf lint...$(COLOR_RESET)'
	@cd configs && buf lint
	@echo '$(COLOR_GREEN)✓ Lint complete$(COLOR_RESET)'

proto-breaking: ## Check for breaking changes against main
	@echo '$(COLOR_BOLD)Checking for breaking changes...$(COLOR_RESET)'
	@cd configs && buf breaking --against ../.git#branch=main || echo '$(COLOR_YELLOW)⚠ Breaking check skipped (requires buf)$(COLOR_RESET)'

proto-build: ## Build descriptor set (image.bin)
	@echo '$(COLOR_BOLD)Building descriptor set...$(COLOR_RESET)'
	@mkdir -p api-docs/descriptors
	@cd configs && buf build -o ../api-docs/descriptors/image.bin || echo '$(COLOR_YELLOW)⚠ Descriptor build skipped (requires buf)$(COLOR_RESET)'
	@echo '$(COLOR_GREEN)✓ Descriptor set built$(COLOR_RESET)'

proto-docs: ## Run the full proto docs pipeline
	@echo '$(COLOR_BOLD)Running full documentation pipeline...$(COLOR_RESET)'
	@go run ./cmd/proto-docs all --config configs/proto-docs.config.yaml
	@echo '$(COLOR_GREEN)✓ Documentation pipeline complete$(COLOR_RESET)'

proto-ci: proto-lint proto-build proto-docs ## Run all proto CI tasks

build-proto-docs: ## Build proto-docs CLI
	@echo '$(COLOR_BOLD)Building proto-docs...$(COLOR_RESET)'
	@mkdir -p $(BUILD_DIR)
	@go build -o $(PROTO_DOCS_BIN) ./cmd/proto-docs
	@echo '$(COLOR_GREEN)✓ proto-docs built: $(PROTO_DOCS_BIN)$(COLOR_RESET)'

build-runtime: ## Build runtime service
	@echo '$(COLOR_BOLD)Building runtime service...$(COLOR_RESET)'
	@mkdir -p $(BUILD_DIR)
	@go build -o $(RUNTIME_BIN) ./cmd/runtime
	@echo '$(COLOR_GREEN)✓ runtime built: $(RUNTIME_BIN)$(COLOR_RESET)'

run-runtime: proto-build build-runtime ## Run the runtime service
	@echo '$(COLOR_BOLD)Starting runtime service...$(COLOR_RESET)'
	@PROTO_DESCRIPTOR_PATH=api-docs/descriptors/image.bin $(RUNTIME_BIN)

test-protoctx: ## Run ProtoContext tests
	@echo '$(COLOR_BOLD)Running ProtoContext tests...$(COLOR_RESET)'
	@go test -v ./tools/protoctx
	@echo '$(COLOR_GREEN)✓ ProtoContext tests complete$(COLOR_RESET)'

test-pipeline: ## Run Pipeline tests
	@echo '$(COLOR_BOLD)Running Pipeline tests...$(COLOR_RESET)'
	@go test -v ./tools/protodocs/pipeline
	@echo '$(COLOR_GREEN)✓ Pipeline tests complete$(COLOR_RESET)'

build-enricher: ## Build protodocs-enricher CLI
	@echo '$(COLOR_BOLD)Building protodocs-enricher...$(COLOR_RESET)'
	@mkdir -p $(BUILD_DIR)
	@go build -o $(ENRICHER_BIN) ./cmd/protodocs-enricher
	@echo '$(COLOR_GREEN)✓ protodocs-enricher built: $(ENRICHER_BIN)$(COLOR_RESET)'

run-enricher: build-enricher ## Run the enricher (requires model and config)
	@echo '$(COLOR_BOLD)Running enrichment...$(COLOR_RESET)'
	@$(ENRICHER_BIN) \
		--config configs/enricher.config.yaml \
		--input api-docs/model/api-doc-model.json \
		--output api-docs/model/api-doc-model-enriched.json \
		--manifest api-docs/enrichment-manifest.json \
		--tenant default

test-enricher: ## Run Enricher tests
	@echo '$(COLOR_BOLD)Running Enricher tests...$(COLOR_RESET)'
	@go test -v ./tools/protodocs/enricher
	@go test -v ./tools/protodocs/enricher/adapters
	@echo '$(COLOR_GREEN)✓ Enricher tests complete$(COLOR_RESET)'

.DEFAULT_GOAL := help
