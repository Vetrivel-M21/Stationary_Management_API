.PHONY: run build test migrate seed docker-up docker-down

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

test:
	go test -v ./...

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down -v
