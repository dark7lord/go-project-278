BIN := bin/webpapp
GOLANGCI_LINT_VERSION := v2.12.2

build:
	go build -o $(BIN) main.go

lint-install:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

lint-uninstall:
	rm -f $(shell which golangci-lint 2>/dev/null)

lint:
	golangci-lint run

lint-fix:
	golangci-lint run --fix

fmt:
	golangci-lint fmt

test:
	go test -race -coverprofile=coverage.out ./... -v 

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

cover-html: cover
	go tool cover -html=coverage.out

check: test lint build clean
	@echo "✅ All checks passed"

clean:
	rm -rf $(BIN) coverage.out tmp