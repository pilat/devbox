.PHONY: all build tidy vendor mocks fmt lint test test-e2e clean

all: tidy vendor lint test

build:
	go build -o devbox ./cmd/devbox/

mocks:
	mockery

tidy:
	go mod tidy

vendor:
	go mod vendor

fmt:
	golangci-lint fmt

lint:
	golangci-lint run ./...

test:
	go test -race ./internal/...

test-e2e:
	go test -v ./tests/e2e/ -timeout 10m -count=1

clean:
	rm -f devbox
