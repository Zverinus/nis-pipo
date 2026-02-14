BINARY := build/nis-pipo
.PHONY: build run test db migrate-up migrate-down swagger docker-build

build:
	go build -o $(BINARY) ./cmd/app

run: build
	./$(BINARY)

test:
	go test ./...

db:
	docker compose up -d db

migrate-up:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$${DB_DSN:-postgres://postgres:1234@localhost:5432/pipo?sslmode=disable}" up

migrate-down:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$${DB_DSN:-postgres://postgres:1234@localhost:5432/pipo?sslmode=disable}" down

swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g internal/transport/router.go -o internal/transport/docs --parseDependency --parseInternal

docker-build:
	docker build -t nis-pipo:latest .
