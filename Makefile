BIN := bin/webapp

.PHONY: build clean
build:
	go build -o $(BIN) cmd/api/main.go

clean:
	rm -rf $(BIN) coverage.out tmp

# Linters
.PHONY: lint-install lint-uninstall lint fmt
GOLANGCI_LINT_VERSION := v2.12.2

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

# Testing
.PHONY: test cover cover-html check
test:
	go test -race -coverprofile=coverage.out ./... -v 

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

cover-html: cover
	go tool cover -html=coverage.out

check: test lint build clean
	@echo "✅ All checks passed"


# Databases
.PHONY: sqlc-gen goose-up goose-down
-include .env
export

sqlc-gen:
	sqlc generate

goose-up:
	goose -dir migrations postgres $(DATABASE_URL) up

goose-down:
	goose -dir migrations postgres $(DATABASE_URL) down

db-rebuild: goose-down goose-up

# Docker
.PHONY: docker-build docker-run docker-stop docker-clean

IMAGE_NAME := link-shortener

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run:
	docker run -d --name $(IMAGE_NAME) \
		-p 8080:8080 \
		--env-file .env \
		$(IMAGE_NAME)

docker-stop:
	docker stop $(IMAGE_NAME) || true
	docker rm $(IMAGE_NAME) || true

docker-clean: docker-stop
	docker rmi $(IMAGE_NAME) || true
