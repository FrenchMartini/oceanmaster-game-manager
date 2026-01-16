.PHONY: help install build run test lint docker-up docker-down docker-logs clean

install:
	npm install
	npx husky install
	go mod download

build: 
	go build -o bin/game-manager ./cmd/server

run: 
	go run cmd/server/main.go

test: 
	go test ./...

lint:
	./scripts/lint.sh

docker-up: 
	docker-compose up --build

docker-down:
	docker-compose down


clean: 
	rm -rf bin/
	go clean
