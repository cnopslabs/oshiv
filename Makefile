OUT := oshiv
PKG := github.com/cnopslabs/oshiv
VERSION := $(shell git describe --always)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)
OS := $(shell uname -s | awk '{print tolower($0)}')
ARCH := $(shell uname -m)

# Targets
.PHONY: build release clean vet staticcheck install-local compile zip html

build: vet staticcheck install-local

release: clean vet staticcheck compile zip html install-local

clean:
	-@rm -rf website/oshiv/downloads/{mac,windows,linux}/*
	-@rm -f website/index.html

vet:
	@go vet ${PKG_LIST}

staticcheck:
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@staticcheck ./...

install-local:
	GOOS=${OS} GOARCH=${ARCH} go build -v -ldflags="-X main.version=${VERSION}"
	go install -v -ldflags="-X main.version=${VERSION}"

compile:
	@echo "Compiling binaries for multiple platforms..."
	$(foreach os_arch, darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 linux/amd64 linux/arm64, \
		GOOS=$(word 1,$(subst /, ,${os_arch})) GOARCH=$(word 2,$(subst /, ,${os_arch})) \
		go build -v -o website/oshiv/downloads/$(word 1,$(subst /, ,${os_arch}))/$(word 2,$(subst /, ,${os_arch}))/$(OUT)_${VERSION}_$(word 1,$(subst /, ,${os_arch}))_$(word 2,$(subst /, ,${os_arch})) -ldflags="-X main.version=${VERSION}";)

zip:
	@echo "Creating ZIP archives for binaries..."
	find website/oshiv/downloads/ -type f -exec zip -j {}.zip {} \;

html:
	cd website/oshiv && go run renderhtml.go ${VERSION} index.tmpl && cd ..
