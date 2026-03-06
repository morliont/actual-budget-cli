BINARY=actual-cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-X github.com/morliont/actual-budget-cli/internal/version.Version=$(VERSION) -X github.com/morliont/actual-budget-cli/internal/version.Commit=$(COMMIT) -X github.com/morliont/actual-budget-cli/internal/version.Date=$(DATE)"

.PHONY: setup lint test build

setup:
	npm install
	go mod tidy

lint:
	go vet ./...

test:
	go test ./...

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/actual-cli
