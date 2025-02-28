DEFAULT_GOAL := build

.PHONY: build
build:
	@echo "Building binary..."
	@go build -o bin/server cmd/server/main.go

.PHONY: lint
lint:
	@echo "Running linter..."
	@golangci-lint run

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
