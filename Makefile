.PHONY: build test lint clean all

BINARY=sneakernet-sync

build:
	go build -o $(BINARY) ./cmd/synccli

test:
	go test -v -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run

cover: test
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -f $(BINARY) sneakernet-sync-* *.exe coverage.out coverage.html

all: lint test build
