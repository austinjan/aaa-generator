.PHONY: all build run install test fmt vet lint clean build-all build-linux build-darwin build-windows help

# Basic build settings for the generator CLI.
GO       ?= go
CMD_PKG  := ./cmd/generator
BIN      ?= generator
ARGS     ?=

# Version from git tag or commit
GIT_TAG  := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION  := $(GIT_TAG)-$(GIT_COMMIT)
LDFLAGS  := -ldflags "-X main.version=$(VERSION)"

all: build

build:
	@echo "Building $(BIN) version $(VERSION)..."
	@$(GO) build -trimpath $(LDFLAGS) -o $(BIN) $(CMD_PKG)

run:
	@echo "Running $(BIN) with ARGS='$(ARGS)'..."
	@$(GO) run $(CMD_PKG) $(ARGS)

install:
	@echo "Installing $(CMD_PKG) version $(VERSION) into $$GOBIN (or $$GOPATH/bin)..."
	@$(GO) install $(LDFLAGS) $(CMD_PKG)

test:
	@$(GO) test ./...

fmt:
	@$(GO) fmt ./...

vet:
	@$(GO) vet ./...

lint: vet

clean:
	@echo "Removing build artifacts..."
	@$(RM) -f $(BIN) $(BIN).exe
	@$(RM) -rf dist/

build-all: build-linux build-darwin build-windows
	@echo "✅ All platform binaries built in dist/"

build-linux:
	@echo "🐧 Building for Linux (amd64)..."
	@mkdir -p dist
	@GOOS=linux GOARCH=amd64 $(GO) build -trimpath $(LDFLAGS) -o dist/$(BIN)-linux-amd64 $(CMD_PKG)
	@echo "✅ dist/$(BIN)-linux-amd64"

build-darwin:
	@echo "🍎 Building for macOS (amd64 and arm64)..."
	@mkdir -p dist
	@GOOS=darwin GOARCH=amd64 $(GO) build -trimpath $(LDFLAGS) -o dist/$(BIN)-darwin-amd64 $(CMD_PKG)
	@GOOS=darwin GOARCH=arm64 $(GO) build -trimpath $(LDFLAGS) -o dist/$(BIN)-darwin-arm64 $(CMD_PKG)
	@echo "✅ dist/$(BIN)-darwin-amd64"
	@echo "✅ dist/$(BIN)-darwin-arm64"

build-windows:
	@echo "🪟 Building for Windows (amd64)..."
	@mkdir -p dist
	@GOOS=windows GOARCH=amd64 $(GO) build -trimpath $(LDFLAGS) -o dist/$(BIN)-windows-amd64.exe $(CMD_PKG)
	@echo "✅ dist/$(BIN)-windows-amd64.exe"

help:
	@echo "Available targets:"
	@echo "  make build         - Compile the generator binary for current platform"
	@echo "  make build-all     - Build binaries for all platforms (Linux, macOS, Windows)"
	@echo "  make build-linux   - Build for Linux (amd64)"
	@echo "  make build-darwin  - Build for macOS (amd64 + arm64)"
	@echo "  make build-windows - Build for Windows (amd64)"
	@echo "  make run ARGS='...'- Run the generator with optional arguments"
	@echo "  make install       - Install the generator into \$${GOBIN}"
	@echo "  make test          - Run Go unit tests"
	@echo "  make fmt           - Format Go source files"
	@echo "  make vet           - Run go vet checks"
	@echo "  make lint          - Alias for go vet"
	@echo "  make clean         - Remove all build artifacts"
	@echo "  make help          - Show this help message"
