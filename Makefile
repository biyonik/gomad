# Makefile
.PHONY: all build test lint fmt clean

# Default Go compiler
GO := go

# Build flags
LDFLAGS := -ldflags="-s -w"

# Windows için CGO ayarları
ifeq ($(OS),Windows_NT)
    export CGO_ENABLED=1
    export CC=gcc
endif

all: lint test build

## build: Build the application
build:
	$(GO) build $(LDFLAGS) -o bin/gomad ./cmd/gomad

## test: Run all tests
test:
	$(GO) test -v -race -cover ./...

## test-short: Run short tests only
test-short:
	$(GO) test -v -short ./...

## lint: Run linter
lint:
	golangci-lint run ./...

## fmt: Format code
fmt:
	$(GO) fmt ./...
	goimports -w .

## clean: Clean build artifacts
clean:
	rm -rf bin/
	$(GO) clean -testcache

## deps: Download dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

## example-hello: Run hello world example
example-hello:
	$(GO) run ./cmd/examples/hello-world/main.go

## help: Show this help
help:
	@echo "GOMAD Makefile Commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'