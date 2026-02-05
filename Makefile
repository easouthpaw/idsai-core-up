APP=cmd/api/main.go
DB_URL?=postgres://postgres:postgres@localhost:5433/idsai?sslmode=disable

.PHONY: run test test-integration up down migrate-up tools swagger

run:
	DATABASE_URL="$(DB_URL)" go run $(APP)

test:
	go test ./...

test-integration:
	DATABASE_URL="$(DB_URL)" go test -tags=integration ./...

up:
	docker compose up -d

down:
	docker compose down -v

tools:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/swaggo/swag/cmd/swag@latest

migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up

swagger:
	swag init -g cmd/api/main.go -o docs/swagger