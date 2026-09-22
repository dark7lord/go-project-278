# Shared variables
BIN := bin/app
IMAGE_NAME := link-shortener

.DEFAULT_GOAL := all

# Load environment (.env) for db/docker commands
-include .env
export

include make/dev.mk make/lint.mk make/test.mk make/db.mk make/docker.mk

.PHONY: all help
all: deps build ## Everything needed to start working: deps + build

help:
	@echo "Available commands:"
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'