BINARY := build/nis-pipo
DB_DSN ?= postgres://postgres:1234@localhost:5432/pipo?sslmode=disable
CACHE_DIR := .cache
GOCACHE ?= $(CURDIR)/$(CACHE_DIR)/go-build
GOMODCACHE ?= $(CURDIR)/$(CACHE_DIR)/go-mod

export GOCACHE
export GOMODCACHE

.PHONY: deps build run test clean up down restart logs migrate-up migrate-down swagger

prepare-cache:
	@mkdir -p build $(GOCACHE) $(GOMODCACHE)

deps: prepare-cache
	go mod download

build: prepare-cache
	go build -o $(BINARY) ./cmd/app

run: build
	./$(BINARY)

test: prepare-cache
	go test ./...

clean:
	rm -f $(BINARY)

up:
	docker compose up -d --build
	@echo "Frontend:   http://localhost:8081"
	@echo "Swagger:    http://localhost:8080/swagger/index.html"
	@echo "Metrics:    http://localhost:8080/metrics"
	@echo "Prometheus: http://localhost:9090"

down:
	docker compose down

restart:
	docker compose down
	docker compose up -d --build
	@echo "Frontend:   http://localhost:8081"
	@echo "Swagger:    http://localhost:8080/swagger/index.html"
	@echo "Metrics:    http://localhost:8080/metrics"
	@echo "Prometheus: http://localhost:9090"

logs:
	docker compose logs -f

migrate-up: prepare-cache
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$(DB_DSN)" up

migrate-down: prepare-cache
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$(DB_DSN)" down

swagger: prepare-cache
	go run github.com/swaggo/swag/cmd/swag@latest init -g internal/transport/router.go -o internal/transport/docs --parseDependency --parseInternal
