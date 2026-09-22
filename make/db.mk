# Databases
.PHONY: sqlc-gen goose-status goose-up goose-down goose-validate db-redo
GOOSE := goose -dir db/migrations postgres $(DATABASE_URL)

sqlc-gen: ## Generate code from SQL (sqlc)
	sqlc generate

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