##@ Development
.PHONY: help setup preflight dev run run-front build deps check clean

help: ## Show this help (also what a bare `make` does)
	@awk 'BEGIN {FS = ":.*## "} \
		/^##@/ {printf "\n\033[1m%s\033[0m\n", substr($$0, 5)} \
		/^[a-zA-Z_-]+:.*## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# A sub-make, so the .env created here is read before db-migrate needs DATABASE_URL
setup: preflight ## After a clone: .env, tools, deps, PostgreSQL, migrations
	@test -f .env || cp .env.example .env
	@$(MAKE) --no-print-directory tools deps db-up db-migrate
	@echo "Ready: make dev"

preflight: ## Check what setup needs: go, npm, a running docker, a free PostgreSQL port
	@scripts/preflight.sh

dev: db-up ## PostgreSQL, the API restarted by air on save, the dashboard
	npx concurrently --kill-others --names api,front "$(GOTOOL) air" "npm run dev:front"

run: build db-up ## PostgreSQL, the built API and the dashboard, no restarts
	npm run dev

run-front: ## The dashboard alone
	npm run dev:front

build: ## Build the API binary to bin/app
	go build -o $(BIN) .

deps: ## Install Go modules and the frontend package
	go mod download
	npm ci

check: test lint build clean ## Everything CI checks: test + lint + build
	@echo "✅ All checks passed"

clean: ## Remove build and test artifacts
	rm -rf $(BIN) coverage*.out tmp
