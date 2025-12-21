.PHONY: all tidy vendor lint test test-e2e

all: tidy vendor lint test

tidy:
	go mod tidy

vendor:
	go mod vendor

lint:
	golangci-lint run

test:
	go test -race ./internal/...

test-e2e:
	go test -v ./tests/e2e/ -timeout 10m
