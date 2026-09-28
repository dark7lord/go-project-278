##@ API docs
.PHONY: api-lint api-html
REDOCLY := npx --yes @redocly/cli@2.54.3

api-lint: ## Validate openapi/openapi.yaml
	$(REDOCLY) lint openapi/openapi.yaml

api-html: ## Build and open the API docs (tmp/api.html)
	$(REDOCLY) build-docs openapi/openapi.yaml -o tmp/api.html
	open tmp/api.html
