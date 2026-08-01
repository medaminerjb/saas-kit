.PHONY: build run test test-integration lint migrate-up migrate-down migrate-create sqlc sqlc-check docker-up docker-down seed clean keys help

# ─── Variables ─────────────────────────────────────────
BINARY_NAME := saaskit
BUILD_DIR := ./bin
MAIN_PATH := ./cmd/saaskit
DATABASE_URL ?= postgres://saaskit:saaskit@localhost:5432/saaskit?sslmode=disable

# ─── Build ─────────────────────────────────────────────
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: ## Run with hot-reload (requires air)
	@which air > /dev/null 2>&1 || go install github.com/air-verse/air@latest
	air

run-direct: ## Run directly without hot-reload
	go run $(MAIN_PATH)

# ─── Testing ───────────────────────────────────────────
test: ## Run unit tests
	go test -v -race -count=1 -short ./...

test-integration: ## Run integration tests (requires running PostgreSQL)
	go test -v -race -count=1 -run Integration ./...

test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ─── Linting ───────────────────────────────────────────
lint: ## Run linter
	@which golangci-lint > /dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

# ─── Database ──────────────────────────────────────────
migrate-up: ## Run all pending migrations
	@which goose > /dev/null 2>&1 || go install github.com/pressly/goose/v3/cmd/goose@latest
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down: ## Rollback the last migration
	@which goose > /dev/null 2>&1 || go install github.com/pressly/goose/v3/cmd/goose@latest
	goose -dir migrations postgres "$(DATABASE_URL)" down

migrate-status: ## Show migration status
	@which goose > /dev/null 2>&1 || go install github.com/pressly/goose/v3/cmd/goose@latest
	goose -dir migrations postgres "$(DATABASE_URL)" status

migrate-create: ## Create a new migration (usage: make migrate-create NAME=add_foo)
	@which goose > /dev/null 2>&1 || go install github.com/pressly/goose/v3/cmd/goose@latest
	goose -dir migrations create $(NAME) sql

# ─── SQLC ──────────────────────────────────────────────
sqlc: ## Generate Go code from SQL queries
	@which sqlc > /dev/null 2>&1 || go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	sqlc generate -f sqlc/sqlc.yaml

sqlc-check: ## Verify sqlc generated code is up to date
	@which sqlc > /dev/null 2>&1 || go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	sqlc diff -f sqlc/sqlc.yaml

# ─── Keys ──────────────────────────────────────────────
keys: ## Generate development signing keys (RS256)
	@mkdir -p keys
	openssl genpkey -algorithm RSA -out keys/active.pem -pkeyopt rsa_keygen_bits:4096
	openssl rsa -pubout -in keys/active.pem -out keys/active.pub
	@echo "Generated RS256 key pair in keys/"

# ─── Docker ────────────────────────────────────────────
docker-up: ## Start all services with Docker Compose
	docker compose up -d

docker-down: ## Stop all services
	docker compose down

docker-build: ## Build Docker image
	docker compose build

docker-logs: ## Tail Docker Compose logs
	docker compose logs -f

# ─── Utilities ─────────────────────────────────────────
clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR) coverage.out coverage.html tmp/

seed: ## Seed development data
	go run $(MAIN_PATH) seed

# ─── Help ──────────────────────────────────────────────
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
