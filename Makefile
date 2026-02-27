BINARY_NAME := vga-demo
CMD_PATH := ./cmd/demo
BUILD_DIR := build

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

# Default target
.PHONY: all
all: build

# === System dependency check ===
.PHONY: check-deps
check-deps:
	@echo "Checking build dependencies..."
	@command -v go >/dev/null 2>&1 || { echo "ERROR: go is not installed. Install Go 1.21+ from https://go.dev/dl/"; exit 1; }
	@command -v gcc >/dev/null 2>&1 || { echo "ERROR: gcc is not installed (needed for CGo)."; exit 1; }
	@command -v pkg-config >/dev/null 2>&1 || { echo "ERROR: pkg-config is not installed."; exit 1; }
	@pkg-config --exists libxmp 2>/dev/null || { echo "ERROR: libxmp-dev is not installed."; echo "  Ubuntu/Debian: sudo apt install libxmp-dev"; echo "  macOS:         brew install libxmp"; echo "  Fedora:        sudo dnf install libxmp-devel"; exit 1; }
	@pkg-config --exists alsa 2>/dev/null || { echo "ERROR: libasound2-dev is not installed (Linux only, needed for audio)."; echo "  Ubuntu/Debian: sudo apt install libasound2-dev"; echo "  Fedora:        sudo dnf install alsa-lib-devel"; exit 1; }
	@echo "All dependencies found."

# === Native build ===
.PHONY: build
build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# === Platform-specific builds ===
.PHONY: build-linux
build-linux:
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
		go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

.PHONY: build-mac-intel
build-mac-intel:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
		CC=o64-clang \
		go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64"

.PHONY: build-mac-arm
build-mac-arm:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 \
		CC=oa64-clang \
		go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64"

.PHONY: build-windows
build-windows:
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
		CC=x86_64-w64-mingw32-gcc \
		go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_PATH)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe"

# === Build all platforms (requires cross-compilation toolchains) ===
.PHONY: build-all
build-all: build-linux build-mac-intel build-mac-arm build-windows

# === Development ===
.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: run-mod
run-mod: build
	@test -n "$(MOD)" || { echo "Usage: make run-mod MOD=path/to/file.mod"; exit 1; }
	./$(BUILD_DIR)/$(BINARY_NAME) -mod $(MOD)

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

.PHONY: deps
deps:
	go mod tidy
	go mod download

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

# === Help ===
.PHONY: help
help:
	@echo "VGA-GO Demo Engine"
	@echo ""
	@echo "Usage:"
	@echo "  make              Build for current platform (native)"
	@echo "  make run          Build and run (no music)"
	@echo "  make run-mod MOD=file.mod  Build and run with a tracker module"
	@echo "  make check-deps   Verify all build dependencies are installed"
	@echo "  make build-linux  Build for Linux amd64"
	@echo "  make build-mac-intel  Build for macOS amd64 (requires osxcross)"
	@echo "  make build-mac-arm    Build for macOS arm64 (requires osxcross)"
	@echo "  make build-windows    Build for Windows (requires mingw-w64)"
	@echo "  make build-all    Build for all platforms"
	@echo "  make clean        Remove build artifacts"
	@echo "  make deps         Download and tidy dependencies"
	@echo "  make fmt          Format Go source"
	@echo "  make vet          Run go vet"
