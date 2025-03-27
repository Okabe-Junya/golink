DEFAULT_GOAL := build

.PHONY: build
build:
	@echo "Building binary..."
	@go build -o bin/server cmd/server/main.go

.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@gofmt -w .
	@goimports -w .

.PHONY: vendor
vendor:
	@echo "Updating vendor..."
	@go mod vendor
	@go mod tidy
	@go mod verify

.PHONY: test-all
test-all: test-unit test-e2e

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	@go test -v ./... -run TestUnit

.PHONY: test-e2e
test-e2e:
	@echo "Running E2E tests..."
	@LANG=C go test -v ./tests/e2e

.PHONY: run
run:
	@echo "Running server..."
	@go run cmd/server/main.go

.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf bin

.PHONY: build-cleanup
build-cleanup:
	@echo "Building cleanup tool..."
	@go build -o bin/cleanup cmd/cleanup/main.go

.PHONY: cleanup
cleanup: build-cleanup
	@echo "Running cleanup job..."
	@./bin/cleanup

.PHONY: cleanup-dry-run
cleanup-dry-run: build-cleanup
	@echo "Running cleanup job (dry run)..."
	@./bin/cleanup --dry-run

.PHONY: cleanup-with-age
cleanup-with-age: build-cleanup
	@echo "Running cleanup job with custom age..."
	@./bin/cleanup --older-than $(age)

.PHONY: build-migrate
build-migrate:
	@echo "Building migration tool..."
	@go build -o bin/migrate cmd/migrate/main.go

.PHONY: migrate
migrate: build-migrate
	@echo "Running migration..."
	@./bin/migrate $(ARGS)

.PHONY: migrate-create-stats
migrate-create-stats: build-migrate
	@echo "Creating stats collection..."
	@./bin/migrate --create-stats

.PHONY: migrate-expired-links
migrate-expired-links: build-migrate
	@echo "Migrating expired links..."
	@./bin/migrate --migrate-expired

.PHONY: migrate-dry-run
migrate-dry-run: build-migrate
	@echo "Running migration (dry run)..."
	@./bin/migrate --dry-run $(ARGS)

.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build            - Build the server binary"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format code"
	@echo "  vendor           - Update vendor dependencies"
	@echo "  test-all         - Run all tests"
	@echo "  test-unit        - Run unit tests"
	@echo "  test-e2e          - Run E2E tests"
	@echo "  run              - Run server"
	@echo "  clean            - Clean up"
	@echo "  cleanup          - Run cleanup job"
	@echo "  cleanup-dry-run  - Run cleanup job (dry run)"
	@echo "  cleanup-with-age - Run cleanup job with custom age"
	@echo "  migrate          - Run migrations with ARGS"
	@echo "  migrate-create-stats - Create link stats collection"
	@echo "  migrate-expired-links - Migrate expired links"
	@echo "  migrate-dry-run  - Run migrations in dry-run mode"
	@echo "  help             - Show this help message"
