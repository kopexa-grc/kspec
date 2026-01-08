# kspec Makefile
# Run `make help` for available commands

.PHONY: help build install clean test lint lint-fix fmt vet check \
        run dev release snapshot hooks deps tidy schema

# Build variables
BINARY_NAME := kspec
BUILD_DIR := ./bin
MAIN_PKG := ./cmd/kspec
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

# Go variables
GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

# Colors
BLUE := \033[34m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

##@ General

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(BLUE)Usage:$(RESET)\n  make $(GREEN)<target>$(RESET)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

build: ## Build the binary
	@echo "$(BLUE)Building $(BINARY_NAME)...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PKG)
	@echo "$(GREEN)Built $(BUILD_DIR)/$(BINARY_NAME)$(RESET)"

install: build ## Install binary to GOBIN
	@echo "$(BLUE)Installing $(BINARY_NAME) to $(GOBIN)...$(RESET)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOBIN)/$(BINARY_NAME)
	@echo "$(GREEN)Installed to $(GOBIN)/$(BINARY_NAME)$(RESET)"

dev: ## Build and run with dev version
	@go run -ldflags="$(LDFLAGS)" $(MAIN_PKG) $(ARGS)

run: build ## Build and run the binary
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

clean: ## Remove build artifacts
	@echo "$(BLUE)Cleaning...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)Cleaned$(RESET)"

schema: ## Generate JSON schema from Policy struct
	@echo "$(BLUE)Generating JSON schema...$(RESET)"
	@go run ./cmd/schema-gen
	@echo "$(GREEN)Schema generated at schema/policy.schema.json$(RESET)"

##@ Testing

test: ## Run tests
	@echo "$(BLUE)Running tests...$(RESET)"
	go test -v -race ./...

test-coverage: ## Run tests with coverage
	@echo "$(BLUE)Running tests with coverage...$(RESET)"
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(RESET)"

test-short: ## Run short tests only
	go test -v -short ./...

##@ Code Quality

lint: ## Run linter
	@echo "$(BLUE)Running linter...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "$(RED)golangci-lint not installed. Run: make lint-install$(RESET)"; \
		exit 1; \
	fi

lint-fix: ## Run linter with auto-fix
	@echo "$(BLUE)Running linter with fixes...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix --timeout=5m; \
	else \
		echo "$(RED)golangci-lint not installed. Run: make lint-install$(RESET)"; \
		exit 1; \
	fi

lint-install: ## Install golangci-lint
	@echo "$(BLUE)Installing golangci-lint...$(RESET)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)Installed golangci-lint$(RESET)"

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(RESET)"
	go fmt ./...
	@echo "$(GREEN)Done$(RESET)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(RESET)"
	go vet ./...

check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)
	@echo "$(GREEN)All checks passed!$(RESET)"

##@ Dependencies

deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(RESET)"
	go mod download

tidy: ## Tidy and verify dependencies
	@echo "$(BLUE)Tidying dependencies...$(RESET)"
	go mod tidy
	go mod verify

##@ Git Hooks

hooks: ## Install git hooks via lefthook
	@echo "$(BLUE)Installing git hooks...$(RESET)"
	@if command -v lefthook >/dev/null 2>&1; then \
		lefthook install; \
		echo "$(GREEN)Git hooks installed$(RESET)"; \
	else \
		echo "$(RED)lefthook not installed. Run: make hooks-install$(RESET)"; \
		exit 1; \
	fi

hooks-install: ## Install lefthook
	@echo "$(BLUE)Installing lefthook...$(RESET)"
	go install github.com/evilmartians/lefthook@latest
	@echo "$(GREEN)Installed lefthook$(RESET)"

hooks-uninstall: ## Uninstall git hooks
	@if command -v lefthook >/dev/null 2>&1; then \
		lefthook uninstall; \
		echo "$(GREEN)Git hooks uninstalled$(RESET)"; \
	fi

##@ Release

release: ## Create a release build (requires goreleaser)
	@echo "$(BLUE)Creating release...$(RESET)"
	goreleaser release --clean

snapshot: ## Create a snapshot release (no publish)
	@echo "$(BLUE)Creating snapshot...$(RESET)"
	goreleaser release --snapshot --clean

release-check: ## Validate goreleaser config
	@echo "$(BLUE)Checking goreleaser config...$(RESET)"
	goreleaser check

##@ Docker (future)

# docker-build: ## Build Docker image
# 	docker build -t kspec:$(VERSION) .

##@ Utilities

version: ## Show version info
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"

tools: lint-install hooks-install ## Install all development tools
	@echo "$(GREEN)All tools installed$(RESET)"

.DEFAULT_GOAL := help
