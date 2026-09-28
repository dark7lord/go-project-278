##@ Docker (the image Render builds)
.PHONY: docker-build docker-up docker-down

docker-build: ## Build the image: API, dashboard, Caddy
	docker compose --profile app build app

docker-up: ## Start the image with PostgreSQL, on http://localhost; migrations run on start
	docker compose --profile app up -d --build --wait

docker-down: ## Stop the image, PostgreSQL keeps running
	docker compose --profile app rm --stop --force app
