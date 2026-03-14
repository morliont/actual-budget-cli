BINARY=actual-cli
GO ?= $(shell command -v go 2>/dev/null || echo /home/node/.openclaw/workspace/.local/go/bin/go)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-X github.com/morliont/actual-budget-cli/internal/version.Version=$(VERSION) -X github.com/morliont/actual-budget-cli/internal/version.Commit=$(COMMIT) -X github.com/morliont/actual-budget-cli/internal/version.Date=$(DATE)"

.PHONY: setup lint test test-workflows build

setup:
	npm install
	$(GO) mod tidy

lint:
	$(GO) vet ./...

test:
	$(GO) test ./...

test-workflows:
	$(GO) test ./internal/app -run '^TestAgenticWorkflow_' -count=1

build:
	$(GO) build $(LDFLAGS) -o bin/$(BINARY) ./cmd/actual-cli
