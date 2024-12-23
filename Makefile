# Application Details
APP_NAME := oshiv
PKG := github.com/cnopslabs/oshiv
VERSION := $(shell git describe --always)
OUTPUT_DIR := build
PLATFORMS := darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 linux/amd64 linux/arm64

# Build Targets
.PHONY: all build release clean vet staticcheck compile zip test check-env

# Default target
all: build

# Build the application for the local environment
build: vet staticcheck install-local

# Release target for multiple platforms
release: clean vet staticcheck compile zip

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf $(OUTPUT_DIR)
	@rm -f website/index.html

# Run `go vet` on the codebase
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run static analysis with staticcheck
staticcheck:
	@echo "Running staticcheck..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@staticcheck ./...

# Detect OS for Windows and set appropriate GOOS and GOARCH
ifeq ($(OS),Windows_NT)
	GOOS_COMPILE := windows
else
	GOOS_COMPILE := $(shell uname -s | tr '[:upper:]' '[:lower:]')
endif

GOARCH_COMPILE := $(shell uname -m | sed -e 's/x86_64/amd64/' -e 's/i[3-6]86/386/')

# Install local binary
install-local:
	@echo "Installing local binary..."
	@mkdir -p $(OUTPUT_DIR) # Ensure the output directory exists
	GOOS=$(GOOS_COMPILE) \
	GOARCH=$(GOARCH_COMPILE) \
	go build -v -ldflags="-X main.version=$(VERSION)" \
	-o $(OUTPUT_DIR)/$(APP_NAME)_$(VERSION)_$(GOOS_COMPILE)_$(GOARCH_COMPILE)

# Compile binaries for multiple platforms
compile:
	@echo "Compiling binaries for multiple platforms..."
	@mkdir -p $(OUTPUT_DIR) # Ensure the output directory exists
	$(foreach platform, $(PLATFORMS), \
		$(eval os_arch = $(subst /, ,$(platform))) \
		GOOS=$(word 1,${os_arch}) GOARCH=$(word 2,${os_arch}) \
		go build -v -ldflags="-X main.version=$(VERSION)" \
		-o $(OUTPUT_DIR)/$(APP_NAME)_$(VERSION)_$(word 1,${os_arch})_$(word 2,${os_arch});)

# Zip compiled binaries
zip:
	@echo "Creating ZIP archives for binaries..."
	find $(OUTPUT_DIR) -type f ! -name "*.zip" -exec zip -j {}.zip {} \;

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Check environment setup
check-env:
	@echo "Checking environment..."
	@command -v go >/dev/null 2>&1 || { echo >&2 "Go is not installed. Aborting."; exit 1; }
	@command -v staticcheck >/dev/null 2>&1 || { echo >&2 "Staticcheck is not installed. Run 'make staticcheck' to install it."; exit 1; }
	@command -v zip >/dev/null 2>&1 || { echo >&2 "Zip is not installed. Aborting."; exit 1; }