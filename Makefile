VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test test-integration clean

build:
	go build $(LDFLAGS) -o infernum ./cmd/infernum

test:
	go test -short ./...

test-integration:
	go test ./...

clean:
	rm -f infernum
