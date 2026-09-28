# Shared variables
BIN := bin/app
IMAGE_NAME := link-shortener
GOBIN := $(shell go env GOPATH)/bin

.DEFAULT_GOAL := all

# Load .env for the goose targets' $(DATABASE_URL) (docker reads .env itself)
-include .env
export

include make/dev.mk make/lint.mk make/test.mk make/db.mk make/docker.mk make/api.mk

.PHONY: all tools help
all: deps build ## Everything needed to start working: deps + build

tools: lint-install sqlc-install goose-install ## Install CLI tools (golangci-lint, sqlc, goose)

help:
	@echo "Available commands:"
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'