.PHONY: build test lint prcm

BINARY = bin/ai-session

build:
	mkdir -p bin
	go build -o $(BINARY) ./cmd/ai-session/

test:
	go test ./...

lint:
	golangci-lint run ./...

prcm: build test lint