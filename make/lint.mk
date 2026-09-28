# Linters
# $(GOBIN), not bare `golangci-lint`: run the pinned binary lint-install put there, not whatever comes first in PATH
.PHONY: lint-install lint lint-fix fmt
GOLANGCI_LINT_VERSION := v2.14.0

# Prebuilt, not `go install` (builds with the local Go toolchain): https://golangci-lint.run/docs/welcome/install/local
lint-install: ## Install golangci-lint (prebuilt binary)
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)

lint: ## Run golangci-lint (expect 0 issues)
	$(GOBIN)/golangci-lint run

lint-fix: ## Run golangci-lint and fix issues
	$(GOBIN)/golangci-lint run --fix

fmt: ## Format code (golangci-lint fmt)
	$(GOBIN)/golangci-lint fmt