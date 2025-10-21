.PHONY: all build run install test fmt vet lint clean help

# Basic build settings for the generator CLI.
GO       ?= go
CMD_PKG  := ./cmd/generator
BIN      ?= generator
ARGS     ?=

all: build

build:
	@echo "Building $(BIN)..."
	@$(GO) build -trimpath -o $(BIN) $(CMD_PKG)

run:
	@echo "Running $(BIN) with ARGS='$(ARGS)'..."
	@$(GO) run $(CMD_PKG) $(ARGS)

install:
	@echo "Installing $(CMD_PKG) into $$GOBIN (or $$GOPATH/bin)..."
	@$(GO) install $(CMD_PKG)

test:
	@$(GO) test ./...

fmt:
	@$(GO) fmt ./...

vet:
	@$(GO) vet ./...

lint: vet

clean:
	@echo "Removing build artifacts..."
	@$(RM) -f $(BIN)

help:
	@echo "Available targets:"
	@echo "  make build    - Compile the generator binary"
	@echo "  make run ARGS='-help' - Run the generator with optional arguments"
	@echo "  make install  - Install the generator into \$${GOBIN}"
	@echo "  make test     - Run Go unit tests"
	@echo "  make fmt      - Format Go source files"
	@echo "  make vet      - Run go vet checks"
	@echo "  make lint     - Alias for go vet"
	@echo "  make clean    - Remove the compiled binary"
	@echo "  make help     - Show this help message"
