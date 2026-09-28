##@ Tests
.PHONY: test test-unit test-integration cover cover-html

test: ## All tests with -race, writes coverage.out (integration ones need Docker)
	go test -race -coverpkg=./... -coverprofile=coverage.out ./...

test-unit: ## Unit tests only, no Docker
	go test -short -race ./...

test-integration: ## internal/app against a PostgreSQL in testcontainers
	go test -race ./internal/app/...

cover: test ## Tests + coverage report
	go tool cover -func=coverage.out

cover-html: test ## Tests + coverage report in the browser
	go tool cover -html=coverage.out
