##@ Database (the compose PostgreSQL)
.PHONY: db-up db-down db-migrate db-rollback db-status db-redo db-reset db-gen
# The migrations embedded in the binary, through the same goose library the app uses
MIGRATE := go run ./cmd/migrate

db-up: ## Start PostgreSQL in Docker, the data lives in the pgdata volume
	docker compose up -d --wait db

db-down: ## Stop PostgreSQL, the data stays
	docker compose stop db

db-migrate: ## Apply the migrations (the app also does on start)
	$(MIGRATE) up

db-rollback: ## Roll back the latest migration
	$(MIGRATE) down

db-status: ## Show the migration status
	$(MIGRATE) status

db-redo: ## Re-run the latest migration
	$(MIGRATE) redo

db-reset: ## Drop all local data and migrate from scratch (asks first)
	@printf "Drop all data of the compose PostgreSQL? [y/N] "; read answer; [ "$$answer" = y ] \
		|| { echo "Cancelled."; exit 1; }
	docker compose down --volumes
	@$(MAKE) --no-print-directory db-up db-migrate

# The sqlc of tools/go.mod reproduces db/generated in git byte for byte.
# sqlc keeps stale files: remove the .sql.go of a renamed or deleted query file first
db-gen: ## Regenerate db/generated from db/queries (sqlc)
	$(GOTOOL) sqlc generate
