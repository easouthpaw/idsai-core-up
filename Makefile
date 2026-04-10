APP ?= cmd/api/main.go
DB_URL ?= postgres://postgres:postgres@localhost:5433/idsai?sslmode=disable
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@latest
REPORT_DIR ?= .tmp/report

.PHONY: run up migrate migrate-all migrate-status test test-integration bench coverage coverage-html report-artifacts

run:
	DATABASE_URL="$(DB_URL)" go run $(APP)

up:
	docker compose up -d postgres minio

migrate:
	$(GOOSE) -dir migrations postgres "$(DB_URL)" up

migrate-all: migrate

migrate-status:
	$(GOOSE) -dir migrations postgres "$(DB_URL)" status

test:
	go test ./...

test-integration:
	DATABASE_URL="$(DB_URL)" go test -tags=integration ./...

bench:
	go test ./internal/services/rbac -run '^$$' -bench . -benchmem -benchtime=2s
	go test ./internal/http/middleware -run '^$$' -bench . -benchmem -benchtime=2s

coverage:
	mkdir -p .tmp
	go test ./... -coverprofile=.tmp/cover.out
	go tool cover -func=.tmp/cover.out

coverage-html:
	mkdir -p .tmp
	go test ./... -coverprofile=.tmp/cover.out
	go tool cover -html=.tmp/cover.out -o .tmp/cover.html

report-artifacts:
	mkdir -p $(REPORT_DIR)
	go test ./... | tee $(REPORT_DIR)/unit.txt
	DATABASE_URL="$(DB_URL)" go test -tags=integration ./... | tee $(REPORT_DIR)/integration.txt
	go test ./internal/services/rbac -run '^$$' -bench . -benchmem -benchtime=2s | tee $(REPORT_DIR)/bench-rbac.txt
	go test ./internal/http/middleware -run '^$$' -bench . -benchmem -benchtime=2s | tee $(REPORT_DIR)/bench-middleware.txt
	go test ./internal/http/handlers -run 'TestProjectFlowHandlerGetMemberAccess' -v | tee $(REPORT_DIR)/security-bola.txt
	go test ./... -coverprofile=$(REPORT_DIR)/cover.out | tee $(REPORT_DIR)/coverage.txt
	go tool cover -func=$(REPORT_DIR)/cover.out | tee $(REPORT_DIR)/coverage-func.txt
