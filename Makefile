APP ?= cmd/api/main.go
DB_URL ?= postgres://postgres:postgres@localhost:5433/idsai?sslmode=disable

.PHONY: run up

run:
	DATABASE_URL="$(DB_URL)" go run $(APP)

up:
	docker compose up -d postgres
