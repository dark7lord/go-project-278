##@ Database (the compose PostgreSQL)
.PHONY: db-up db-down db-migrate db-rollback db-status db-redo db-validate db-reset db-gen
GOOSE := $(GOTOOL) goose -dir db/migrations postgres "$(DATABASE_URL)"

db-up: ## Start PostgreSQL in Docker, the data lives in the pgdata volume
	docker compose up -d --wait db

db-down: ## Stop PostgreSQL, the data stays
	docker compose stop db

db-migrate: ## Apply the migrations
	$(GOOSE) up

db-rollback: ## Roll back the latest migration
	$(GOOSE) down

db-status: ## Show the migration status
	$(GOOSE) status

db-redo: ## Re-run the latest migration
	$(GOOSE) redo

db-validate: ## Check the migration files without running them
	$(GOOSE) validate

db-reset: ## Drop all local data and migrate from scratch (asks first)
	@printf "Drop all data of the compose PostgreSQL? [y/N] "; read answer; [ "$$answer" = y ] \
		|| { echo "Cancelled."; exit 1; }
	docker compose down --volumes
	@$(MAKE) --no-print-directory db-up db-migrate

# The sqlc of tools/go.mod reproduces db/generated in git byte for byte.
# sqlc keeps stale files: remove the .sql.go of a renamed or deleted query file first
db-gen: ## Regenerate db/generated from db/queries (sqlc)
	$(GOTOOL) sqlc generate
