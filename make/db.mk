# Databases
SQLC_VERSION := v1.31.1
# Keep in sync with ARG GOOSE_VERSION in the Dockerfile
GOOSE_VERSION := v3.28.0

.PHONY: sqlc-gen sqlc-install goose-install goose-status goose-up goose-down goose-validate db-redo
GOOSE := $(GOBIN)/goose -dir db/migrations postgres $(DATABASE_URL)

sqlc-install: ## Install sqlc (pinned version)
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)

goose-install: ## Install goose (pinned version, as in the image)
	go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)

# $(GOBIN), not bare `sqlc`: only sqlc's version reproduces db/generated in git
sqlc-gen: ## Generate code from SQL (sqlc)
	$(GOBIN)/sqlc generate

goose-status: ## Show DB migration status
	$(GOOSE) status

goose-up: ## Apply DB migrations
	$(GOOSE) up

goose-down: ## Roll back latest DB migration
	$(GOOSE) down

goose-validate: ## Validate migration files without running them
	$(GOOSE) validate

db-redo: ## Re-run the latest migration (redo)
	$(GOOSE) redo