APP ?= cmd/api/main.go
DB_URL ?= postgres://postgres:postgres@localhost:5433/idsai?sslmode=disable
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@latest

.PHONY: run up migrate migrate-status

run:
	DATABASE_URL="$(DB_URL)" go run $(APP)

up:
	docker compose up -d postgres

migrate:
	$(GOOSE) -dir migrations postgres "$(DB_URL)" up

migrate-status:
	$(GOOSE) -dir migrations postgres "$(DB_URL)" status
