# nis-pipo

build:
	go build -o build/nis-pipo ./cmd/app

run: build
	./build/nis-pipo

test:
	go test ./...

db:
	docker compose up -d db

migrate-up:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$${DB_DSN:-postgres://postgres:1234@localhost:5432/pipo?sslmode=disable}" up

migrate-down:
	go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$${DB_DSN:-postgres://postgres:1234@localhost:5432/pipo?sslmode=disable}" down

docker-build:
	docker build -t nis-pipo:latest .
