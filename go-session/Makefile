.PHONY: build test lint

BINARY = bin/ai-session

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd/ai-session/

test:
	go test ./...

lint:
	go vet ./...
