# Detect OS platform
UNAME_S := $(shell uname -s 2>/dev/null || echo "Windows_NT")
ifeq ($(UNAME_S),Linux)
    OS := Linux
endif
ifeq ($(UNAME_S),Darwin)
    OS := Darwin
endif
ifeq ($(OS),)
    OS := Windows_NT
endif

# Binary name
BINARY_NAME = nvrhikvision-exporter
ifeq ($(OS),Windows_NT)
    BINARY_NAME := nvrhikvision-exporter.exe
endif

# Variables
VERSION ?= 0.1.0
BUILD_DIR := ./dist
SCRIPTS_DIR := scripts

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOFMT := $(GOCMD) fmt
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Cross-platform file operations
ifeq ($(OS),Windows_NT)
    RM := powershell -Command "if (Test-Path '$(BUILD_DIR)') { Remove-Item -Recurse -Force '$(BUILD_DIR)' }"
    MKDIR := powershell -Command "if (-not (Test-Path '$(BUILD_DIR)')) { New-Item -ItemType Directory -Force -Path '$(BUILD_DIR)' | Out-Null }"
else
	RM := rm -rf $(BUILD_DIR)
	MKDIR := mkdir -p $(BUILD_DIR)
endif

.PHONY: help build build-all build-linux build-windows build-darwin run run-config clean test fmt lint deps docker-build docker-run mkdir

help:
	@echo "Hikvision NVR Exporter - Available Commands"
	@echo ""
	@echo "Platform: $(OS)"
	@echo ""
	@echo "Usage:"
	@echo "  make build        Build the exporter binary"
	@echo "  make build-all    Build all platform binaries via scripts/"
	@echo "  make build-linux  Build Linux binaries via scripts/"
	@echo "  make build-windows Build Windows binaries via scripts/"
	@echo "  make build-darwin Build macOS binaries via scripts/"
	@echo "  make run          Build and run the exporter"
	@echo "  make run-config   Run with specific config file (CONFIG=path/to/config.yaml)"
	@echo "  make clean        Remove build artifacts"
	@echo "  make test         Run tests"
	@echo "  make fmt          Format code"
	@echo "  make lint         Run linter (requires golangci-lint)"
	@echo "  make deps         Download and tidy dependencies"
	@echo "  make docker-build Build Docker image (requires Docker)"
	@echo "  make docker-run   Run Docker container (requires Docker)"
	@echo ""

build: mkdir
	@echo "[$(OS)] Building $(BINARY_NAME) v$(VERSION)..."
	@cd cmd/exporter && $(GOBUILD) -o ../../$(BUILD_DIR)/$(BINARY_NAME) -ldflags="-X main.Version=$(VERSION)"
	@echo "[$(OS)] Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-all:
ifeq ($(OS),Windows_NT)
	@powershell -ExecutionPolicy Bypass -File .\$(SCRIPTS_DIR)\build.ps1 -Version $(VERSION) -Target all
else
	@bash ./$(SCRIPTS_DIR)/build.sh $(VERSION) all
endif

build-linux:
ifeq ($(OS),Windows_NT)
	@powershell -ExecutionPolicy Bypass -File .\$(SCRIPTS_DIR)\build.ps1 -Version $(VERSION) -Target linux/amd64
else
	@bash ./$(SCRIPTS_DIR)/build.sh $(VERSION) linux/amd64
endif

build-windows:
ifeq ($(OS),Windows_NT)
	@powershell -ExecutionPolicy Bypass -File .\$(SCRIPTS_DIR)\build.ps1 -Version $(VERSION) -Target windows/amd64
else
	@bash ./$(SCRIPTS_DIR)/build.sh $(VERSION) windows/amd64
endif

build-darwin:
ifeq ($(OS),Windows_NT)
	@powershell -ExecutionPolicy Bypass -File .\$(SCRIPTS_DIR)\build.ps1 -Version $(VERSION) -Target darwin/arm64
else
	@bash ./$(SCRIPTS_DIR)/build.sh $(VERSION) darwin/arm64
endif

mkdir:
	@$(MKDIR)

run: build
	@echo "[$(OS)] Running $(BINARY_NAME)..."
	@cd $(BUILD_DIR) && ./$(BINARY_NAME) -config=../config.yaml

run-config: build
ifdef CONFIG
	@echo "[$(OS)] Running $(BINARY_NAME) with config: $(CONFIG)"
	@cd $(BUILD_DIR) && ./$(BINARY_NAME) -config=../$(CONFIG)
else
	@echo "Error: CONFIG not specified. Usage: make run-config CONFIG=path/to/config.yaml"
	@exit 1
endif

clean:
	@echo "[$(OS)] Cleaning up..."
	@$(RM)
	@$(GOCLEAN)
	@echo "[$(OS)] Clean complete"

test:
	@echo "[$(OS)] Running tests..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-coverage: test
	@echo "[$(OS)] Generating coverage report..."
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
ifeq ($(OS),Windows_NT)
	@start coverage.html
else
	@open coverage.html
endif

fmt:
	@echo "[$(OS)] Formatting code..."
	@$(GOFMT) ./...
	@echo "[$(OS)] Format complete"

lint:
	@echo "[$(OS)] Running linter..."
	@golangci-lint run ./...

deps:
	@echo "[$(OS)] Downloading dependencies..."
	@$(GOGET) -u ./...
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "[$(OS)] Dependencies updated"

docker-build:
	@echo "[$(OS)] Building Docker image..."
	@docker build -f deploy/Dockerfile -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .
	@echo "[$(OS)] Docker build complete"

docker-run:
	@echo "[$(OS)] Running Docker container..."
	@docker run --rm -p 9102:9102 $(BINARY_NAME):latest

docker-compose-up:
	@echo "[$(OS)] Starting Docker Compose services..."
	@docker-compose -f deploy/docker-compose.yml up

docker-compose-down:
	@echo "[$(OS)] Stopping Docker Compose services..."
	@docker-compose -f deploy/docker-compose.yml down

# Development helpers
install-tools:
	@echo "[$(OS)] Installing development tools..."
	@$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint
	@echo "[$(OS)] Tools installed"
