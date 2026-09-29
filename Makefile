# Shared variables
BIN := bin/app
GOBIN := $(shell go env GOPATH)/bin
# sqlc, air and mockery are pinned in tools/go.mod, apart from the app's dependencies
GOTOOL := go tool -modfile=tools/go.mod

.DEFAULT_GOAL := help

# Load .env, so $(DATABASE_URL) and friends reach the recipes and the app
-include .env
export

# The order of the files is the order of the sections in `make help`
include make/dev.mk make/test.mk make/lint.mk make/db.mk make/docker.mk make/api.mk make/tools.mk
