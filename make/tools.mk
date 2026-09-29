##@ Tools
.PHONY: tools lint-install

# sqlc, goose, air and mockery: versions in tools/go.mod, run as $(GOTOOL) <name>.
# golangci-lint stays a prebuilt binary, as its authors recommend:
# https://golangci-lint.run/docs/welcome/install/local
GOLANGCI_LINT_VERSION := v2.14.0

tools: lint-install ## Install golangci-lint, build sqlc, goose, air, mockery from tools/go.mod
	go build -modfile=tools/go.mod tool

# Skipped when the pinned version is already in $(GOBIN)
lint-install:
	@$(GOBIN)/golangci-lint version --short 2>/dev/null | grep -qx '$(GOLANGCI_LINT_VERSION:v%=%)' \
		|| curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)
