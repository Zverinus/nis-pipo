BINARY := build/nis-pipo
DB_DSN ?= postgres://postgres:1234@localhost:5432/pipo?sslmode=disable

.PHONY: build run test clean force-build db migrate-up migrate-down swagger docker-build

build:
	@mkdir -p build
	go build -o $(BINARY) ./cmd/app

force-build:
	@mkdir -p build
	go build -a -o $(BINARY) ./cmd/app

run: build
	./$(BINARY)

test:
	go test ./...

clean:
	rm -f $(BINARY)

db:
	docker compose up -d db

migrate-up:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$(DB_DSN)" up

migrate-down:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$(DB_DSN)" down

swagger:
	go run github.com/swaggo/swag/cmd/swag@latest init -g internal/transport/router.go -o internal/transport/docs --parseDependency --parseInternal

docker-build:
	docker build -t nis-pipo:latest .
