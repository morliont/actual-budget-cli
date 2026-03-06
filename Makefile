BINARY := actual-cli
MODULE := github.com/morliont/actual-budget-cli
DIST_DIR := dist

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Deterministic/reproducible-friendly flags.
GOFLAGS := -trimpath
LDFLAGS := -s -w -X $(MODULE)/internal/version.Version=$(VERSION) -X $(MODULE)/internal/version.Commit=$(COMMIT) -X $(MODULE)/internal/version.Date=$(DATE)

PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64

.PHONY: setup lint test fmt-check build clean release-artifacts release-notes

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
	@set -e; \
	for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		out="$(DIST_DIR)/$(BINARY)_$(VERSION)_$${GOOS}_$${GOARCH}"; \
		if [ "$$GOOS" = "windows" ]; then out="$$out.exe"; fi; \
		echo "building $$out"; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o "$$out" ./cmd/actual-cli; \
	done
	@(cd $(DIST_DIR) && sha256sum * > checksums.txt)

release-notes:
	./scripts/release-notes.sh > $(DIST_DIR)/RELEASE_NOTES.md
