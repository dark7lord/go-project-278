# Docker
.PHONY: docker-build docker-run docker-stop docker-clean

docker-build: ## Build the image
	docker build -t $(IMAGE_NAME) .

docker-run: ## Run the container
	docker run -d --name $(IMAGE_NAME) \
		-p 8080:8080 \
		-p 80:80 \
		--env-file .env \
		$(IMAGE_NAME)

docker-stop: ## Stop and remove the container
	docker stop $(IMAGE_NAME) || true
	docker rm $(IMAGE_NAME) || true

docker-clean: docker-stop ## Stop the container and remove the image
	docker rmi $(IMAGE_NAME) || true