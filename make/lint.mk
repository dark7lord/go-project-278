##@ Code quality
# $(GOBIN), not bare `golangci-lint`: the pinned binary from `make tools`, not whatever comes first in PATH
.PHONY: lint lint-fix fmt

# config verify first: CI's action checks the config against the schema, run alone does not
lint: ## golangci-lint, expects 0 issues
	$(GOBIN)/golangci-lint config verify
	$(GOBIN)/golangci-lint run

lint-fix: ## golangci-lint with fixes applied
	$(GOBIN)/golangci-lint run --fix

fmt: ## Format the code
	$(GOBIN)/golangci-lint fmt
