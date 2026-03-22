.PHONY: build run docker-up docker-down tidy

build:
	go build -o bin/bot ./cmd/bot

run:
	go run ./cmd/bot

tidy:
	go mod tidy

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down
