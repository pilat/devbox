.PHONY: all tidy vendor lint test test-e2e build clean

all: tidy vendor lint test build

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

build:
	go build -o devbox ./cmd/devbox/

clean:
	rm -f devbox
