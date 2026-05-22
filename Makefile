APP ?= cmd/api/main.go
DB_URL ?= postgres://postgres:postgres@localhost:5433/idsai?sslmode=disable
TEST_DB_URL ?= postgres://postgres:postgres@localhost:5433/idsai_test?sslmode=disable
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@latest
REPORT_DIR ?= .tmp/report

.PHONY: run up migrate migrate-all migrate-status prepare-test-db test test-integration bench coverage covarege coverage-html report-artifacts

run:
	set -a && . ./.env && set +a && go run $(APP)

up:
	docker compose up -d postgres minio redis

migrate:
	$(GOOSE) -dir migrations postgres "$(DB_URL)" up

migrate-all: migrate

migrate-status:
	$(GOOSE) -dir migrations postgres "$(DB_URL)" status

prepare-test-db:
	docker compose up -d postgres
	@until docker compose exec -T postgres pg_isready -U postgres -d postgres >/dev/null 2>&1; do sleep 1; done
	@docker compose exec -T postgres psql -U postgres -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'idsai_test'" | grep -q 1 || docker compose exec -T postgres createdb -U postgres idsai_test
	$(GOOSE) -dir migrations postgres "$(TEST_DB_URL)" up

test:
	@start=$$(date +%s); \
	log=$$(mktemp); \
	if go test ./... > "$$log" 2>&1; then \
		packages=$$(awk '/^(ok|\?)[[:space:]]/ { count++ } END { print count + 0 }' "$$log"); \
		elapsed=$$(( $$(date +%s) - start )); \
		printf '\n'; \
		printf '  TEST RESULT\n'; \
		printf '  -----------\n'; \
		printf '  Status:   PASS\n'; \
		printf '  Packages: %s\n' "$$packages"; \
		printf '  Time:     %ss\n\n' "$$elapsed"; \
	else \
		elapsed=$$(( $$(date +%s) - start )); \
		printf '\n'; \
		printf '  TEST RESULT\n'; \
		printf '  -----------\n'; \
		printf '  Status:   FAIL\n'; \
		printf '  Time:     %ss\n\n' "$$elapsed"; \
		printf '  Last output:\n'; \
		tail -40 "$$log"; \
		rm -f "$$log"; \
		exit 1; \
	fi; \
	rm -f "$$log"

test-integration: prepare-test-db
	DATABASE_URL="$(TEST_DB_URL)" go test -tags=integration ./...

bench:
	@mkdir -p .tmp; \
	start=$$(date +%s); \
	log=.tmp/bench.log; \
	: > "$$log"; \
	if go test ./internal/services/rbac -run '^$$' -bench . -benchmem -benchtime=2s >> "$$log" 2>&1 && \
		go test ./internal/http/middleware -run '^$$' -bench . -benchmem -benchtime=2s >> "$$log" 2>&1; then \
		elapsed=$$(( $$(date +%s) - start )); \
		printf '\n'; \
		printf '  BENCH RESULT\n'; \
		printf '  ------------\n'; \
		printf '  Status: PASS\n'; \
		printf '  Time:   %ss\n\n' "$$elapsed"; \
		awk '/^Benchmark/ { \
			name=$$1; sub(/-[0-9]+$$/, "", name); \
			printf "  %-56s %10s ns/op %8s B/op %6s allocs/op\n", name, $$3, $$5, $$7 \
		}' "$$log"; \
		printf '\n'; \
	else \
		elapsed=$$(( $$(date +%s) - start )); \
		printf '\n'; \
		printf '  BENCH RESULT\n'; \
		printf '  ------------\n'; \
		printf '  Status: FAIL\n'; \
		printf '  Time:   %ss\n\n' "$$elapsed"; \
		printf '  Last output:\n'; \
		tail -40 "$$log"; \
		exit 1; \
	fi

coverage:
	@mkdir -p .tmp; \
	start=$$(date +%s); \
	log=.tmp/coverage-test.log; \
	func=.tmp/coverage-func.txt; \
	if go test ./... -coverprofile=.tmp/cover.out > "$$log" 2>&1; then \
		go tool cover -func=.tmp/cover.out > "$$func"; \
		total=$$(awk '/^total:/ { print $$3 }' "$$func"); \
		packages=$$(awk '/^(ok|\?)[[:space:]]/ { count++ } END { print count + 0 }' "$$log"); \
		elapsed=$$(( $$(date +%s) - start )); \
		printf '\n'; \
		printf '  COVERAGE RESULT\n'; \
		printf '  ---------------\n'; \
		printf '  Status:   PASS\n'; \
		printf '  Total:    %s\n' "$$total"; \
		printf '  Packages: %s\n' "$$packages"; \
		printf '  Time:     %ss\n' "$$elapsed"; \
		printf '  Profile:  .tmp/cover.out\n\n'; \
	else \
		elapsed=$$(( $$(date +%s) - start )); \
		printf '\n'; \
		printf '  COVERAGE RESULT\n'; \
		printf '  ---------------\n'; \
		printf '  Status:   FAIL\n'; \
		printf '  Time:     %ss\n\n' "$$elapsed"; \
		printf '  Last output:\n'; \
		tail -40 "$$log"; \
		exit 1; \
	fi

covarege: coverage

coverage-html:
	mkdir -p .tmp
	go test ./... -coverprofile=.tmp/cover.out
	go tool cover -html=.tmp/cover.out -o .tmp/cover.html

report-artifacts:
	mkdir -p $(REPORT_DIR)
	go test ./... | tee $(REPORT_DIR)/unit.txt
	$(MAKE) prepare-test-db
	DATABASE_URL="$(TEST_DB_URL)" go test -tags=integration ./... | tee $(REPORT_DIR)/integration.txt
	go test ./internal/services/rbac -run '^$$' -bench . -benchmem -benchtime=2s | tee $(REPORT_DIR)/bench-rbac.txt
	go test ./internal/http/middleware -run '^$$' -bench . -benchmem -benchtime=2s | tee $(REPORT_DIR)/bench-middleware.txt
	go test ./internal/http/handlers -run 'TestProjectFlowHandlerGetMemberAccess' -v | tee $(REPORT_DIR)/security-bola.txt
	go test ./... -coverprofile=$(REPORT_DIR)/cover.out | tee $(REPORT_DIR)/coverage.txt
	go tool cover -func=$(REPORT_DIR)/cover.out | tee $(REPORT_DIR)/coverage-func.txt
