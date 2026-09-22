# Linters
.PHONY: lint-install lint-uninstall lint lint-fix fmt
GOLANGCI_LINT_VERSION := v2.12.2

lint-install: ## Install golangci-lint at pinned version
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint-uninstall: ## Remove golangci-lint
	rm -f $(shell which golangci-lint 2>/dev/null)

lint: ## Run golangci-lint (expect 0 issues)
	golangci-lint run

lint-fix: ## Run golangci-lint and fix issues
	golangci-lint run --fix

fmt: ## Format code (golangci-lint fmt)
	golangci-lint fmt