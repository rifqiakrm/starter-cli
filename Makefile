# ==============================================================================
# LINTING AND CODE QUALITY TARGETS
# ==============================================================================

GO_FILES = $(shell find . -name "*.go" -not -path "./test/integration/*" -not -path "./features/*")

.PHONY: lint
lint: clean-lint-cache ## Run all linters
	@echo "🔍 Running golangci-lint..."
	@golangci-lint run ./...

.PHONY: lint-fix
lint-fix: clean-lint-cache ## Run linters and auto-fix what's possible
	@echo "🔧 Running golangci-lint with fixes..."
	@golangci-lint run --fix ./...

.PHONY: clean-lint-cache
clean-lint-cache: ## Clean the golangci-lint cache
	@echo "🧹 Cleaning lint cache..."
	@golangci-lint cache clean

.PHONY: format
format: ## Format Go code
	@echo "💅 Formatting code..."
	@cmd/bin/format.sh

.PHONY: check-imports
check-imports: ## Check import ordering
	@echo "📦 Checking imports..."
	@cmd/bin/check-import.sh

.PHONY: vet
vet: ## Run go vet
	@echo "🔎 Running go vet..."
	@go vet ./...

.PHONY: staticcheck
staticcheck: ## Run staticcheck
	@echo "📋 Running staticcheck..."
	@staticcheck ./...

.PHONY: tidy
tidy: ## Tidy Go modules
	@echo "🧹 Tidying modules..."
	@go mod tidy

.PHONY: pretty
pretty: tidy format lint ## Run all code quality checks (tidy, format, lint)

.PHONY: validate-migration
validate-migration: ## Validate database migrations
	@echo "🗃️ Validating migrations..."
	@cmd/bin/validate-migration.sh

# ==============================================================================
# QUALITY REPORT TARGETS
# ==============================================================================

.PHONY: lint-report
lint-report: clean-lint-cache ## Generate lint report
	@echo "📊 Generating lint report..."
	@golangci-lint run ./... --out-format=checkstyle > lint-report.xml

.PHONY: code-quality
code-quality: pretty vet staticcheck ## Comprehensive code quality check

# ==============================================================================
# HELP
# ==============================================================================

.PHONY: help
help: ## Show this help message
	@echo "Linting and Code Quality Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help