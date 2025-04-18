APP_NAME := go-app
DOCKER_IMAGE := $(APP_NAME):latest

.PHONY: all build clean run stop logs help

all: build

build:
	docker build --no-cache -t $(DOCKER_IMAGE) .

clean:
	docker rmi -f $(DOCKER_IMAGE)

run:
	docker run -d --name $(APP_NAME) -p 8080:8080 $(DOCKER_IMAGE)

stop:
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true

logs:
	docker logs -f $(APP_NAME)

swagger: ## Generate swagger documentation
	@swag init -g cmd/server/main.go -o docs/api --outputTypes go,yaml --parseDependency --parseInternal

help:
	@echo "Available commands:"
	@echo "  make build    - Build the Docker image without cache"
	@echo "  make run      - Run the Docker container"
	@echo "  make stop     - Stop and remove the Docker container"
	@echo "  make clean    - Remove the Docker image"
	@echo "  make logs     - View container logs"
