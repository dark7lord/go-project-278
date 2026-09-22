# Development
.PHONY: build clean run run-front deps
build: ## Build the app binary
	go build -o $(BIN) .

clean: ## Remove build and test artifacts
	rm -rf $(BIN) coverage*.out tmp

run: build ## Build and run the server and frontend
	npm run dev

run-front: ## Run the frontend in dev mode
	npm run dev:front

deps: ## Install backend and frontend dependencies
	go mod download
	npm ci