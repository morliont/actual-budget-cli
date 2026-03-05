BINARY=actual-cli

.PHONY: setup lint test build

setup:
	npm install
	go mod tidy

lint:
	go vet ./...

test:
	go test ./...

build:
	go build -o bin/$(BINARY) ./cmd/actual-cli
