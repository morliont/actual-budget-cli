BINARY := actual-cli
MODULE := github.com/morliont/actual-budget-cli
DIST_DIR := dist
GORELEASER ?= $(shell command -v goreleaser 2>/dev/null || echo $(HOME)/go/bin/goreleaser)

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Deterministic/reproducible-friendly flags.
GOFLAGS := -trimpath
LDFLAGS := -s -w -X $(MODULE)/internal/version.Version=$(VERSION) -X $(MODULE)/internal/version.Commit=$(COMMIT) -X $(MODULE)/internal/version.Date=$(DATE)

.PHONY: setup lint test fmt-check build clean release-artifacts release-notes goreleaser-check goreleaser-dry-run

setup:
	npm ci
	go mod tidy

lint:
	go vet ./...

test:
	go test ./...

fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "gofmt check failed. Run: gofmt -w ." && gofmt -l . && exit 1)

build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/actual-cli

clean:
	rm -rf $(DIST_DIR)

release-artifacts: clean
	$(GORELEASER) release --snapshot --clean --skip=publish --skip=announce

release-notes:
	./scripts/release-notes.sh > $(DIST_DIR)/RELEASE_NOTES.md

goreleaser-check:
	$(GORELEASER) check

goreleaser-dry-run: clean
	$(GORELEASER) release --snapshot --clean --skip=publish --skip=announce
