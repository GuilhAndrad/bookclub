.PHONY: run build tidy docker-up docker-down docker-up-all

run:
	go run ./cmd/api/main.go

build:
	go build -o bin/bookclub-api ./cmd/api/main.go

tidy:
	go mod tidy

docker-up:
	docker compose -f build/docker-compose.yml up -d postgres

docker-up-all:
	docker compose -f build/docker-compose.yml up -d

docker-down:
	docker compose -f build/docker-compose.yml down