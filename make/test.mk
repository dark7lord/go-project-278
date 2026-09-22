# Testing
.PHONY: test test-unit test-integration cover cover-html check

test: ## Run all tests (with Docker) and write coverage.out
	go test -race -coverpkg=./... -coverprofile=coverage.out ./...

test-unit: ## Run unit tests only (fast, no Docker)
	go test -short -race ./...

test-integration: ## Run integration tests for internal/app (testcontainers)
	go test -race ./internal/app/...

cover: test ## Run all tests and print coverage report
	go tool cover -func=coverage.out

cover-html: test ## Run all tests and open HTML coverage report
	go tool cover -html=coverage.out

check: test lint build clean ## Run all checks: test + lint + build + clean
	@echo "✅ All checks passed"