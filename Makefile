.PHONY: help install build run test lint docker-up docker-down docker-logs clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies and setup Husky
	npm install
	npx husky install
	go mod download

build: ## Build the application
	go build -o bin/game-manager ./cmd/server

run: ## Run the application locally
	go run cmd/server/main.go

test: ## Run tests
	go test ./...

lint: ## Run linter
	./scripts/lint.sh

docker-up: ## Start all services with Docker Compose
	docker-compose up --build

docker-down: ## Stop all services
	docker-compose down

docker-logs: ## View Docker Compose logs
	docker-compose logs -f

clean: ## Clean build artifacts
	rm -rf bin/
	go clean
