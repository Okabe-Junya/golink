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

.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

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

.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build  - Build the binary"
	@echo "  lint   - Run linter"
	@echo "  test   - Run tests"
	@echo "  run    - Run server"
	@echo "  clean  - Clean up"
	@echo "  help   - Show this help message"
