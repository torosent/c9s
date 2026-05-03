.PHONY: all build run test test-race lint fmt fmt-check coverage ci clean install-tools

SHELL        := /bin/bash
GO          ?= go
BIN          := bin/c9s
PKG          := ./...
COVER_OUT    := coverage.out

all: build

build:
	$(GO) build -o $(BIN) ./cmd/c9s

run:
	$(GO) run ./cmd/c9s

test:
	$(GO) test -count=1 $(PKG)

test-race:
	$(GO) test -count=1 -race -coverprofile=$(COVER_OUT) -covermode=atomic $(PKG)

coverage: test-race
	$(GO) tool cover -func=$(COVER_OUT) | tail -n 1

lint:
	golangci-lint run $(PKG)

fmt:
	gofumpt -w .

fmt-check:
	@diff -u <(echo -n) <(gofumpt -d .) || (echo "run 'make fmt'"; exit 1)

ci: fmt-check lint test-race coverage

clean:
	rm -rf bin/ $(COVER_OUT) coverage.html

install-tools:
	$(GO) install mvdan.cc/gofumpt@latest
	$(GO) install honnef.co/go/tools/cmd/staticcheck@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest


.PHONY: docs-hotkeys
docs-hotkeys:
	@echo "Generating docs/hotkeys.md..."
	@go run ./cmd/gen-hotkeys > docs/hotkeys.md
	@echo "✓ docs/hotkeys.md generated"
