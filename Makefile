BINARY ?= tmgc

.PHONY: build test

build:
	mkdir -p ./dist
	go build -o ./dist/$(BINARY) ./cmd/tmgc

test:
	go test ./...
