BINARY ?= tmgc
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: build test

build:
	mkdir -p ./dist
	go build -ldflags "-X main.version=$(VERSION)" -o ./dist/$(BINARY) ./cmd/tmgc

test:
	go test ./...
