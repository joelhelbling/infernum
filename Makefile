VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test test-integration clean

build:
	go build $(LDFLAGS) -o ollama-bench ./cmd/ollama-bench

test:
	go test -short ./...

test-integration:
	go test ./...

clean:
	rm -f ollama-bench
