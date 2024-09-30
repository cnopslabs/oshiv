OUT := oshiv
PKG := github.com/dnlloyd/oshiv
VERSION := $(shell git describe --always)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)
OS := $(shell uname -s | awk '{print tolower($0)}')
ARCH := $(shell uname -m)

build: vet staticcheck install-local

release: clean vet staticcheck compile zip html install-local

clean:
	-@rm -fr website/oshiv/downloads/mac/intel/${OUT}*
	-@rm -fr website/oshiv/downloads/mac/arm/${OUT}*
	-@rm -fr website/oshiv/downloads/windows/intel/${OUT}*
	-@rm -fr website/oshiv/downloads/windows/arm/${OUT}*
	-@rm -fr website/oshiv/downloads/linux/intel/${OUT}*
	-@rm -fr website/oshiv/downloads/linux/arm/${OUT}*
	-@rm -f website/index.html

vet:
	@go vet ${PKG_LIST}

staticcheck:
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@staticcheck ./...

# test:
# 	@go test -short ${PKG_LIST}

install-local:
	GOOS=${OS} GOARCH=${ARCH} go build -v -ldflags="-X main.version=${VERSION}"
	go install -v -ldflags="-X main.version=${VERSION}"

compile:
	GOOS=darwin GOARCH=amd64 go build -v -o website/oshiv/downloads/mac/intel/${OUT}_${VERSION}_darwin_amd64 -ldflags="-X main.version=${VERSION}"
	GOOS=darwin GOARCH=arm64 go build -v -o website/oshiv/downloads/mac/arm/${OUT}_${VERSION}_darwin_arm64 -ldflags="-X main.version=${VERSION}"
	GOOS=windows GOARCH=amd64 go build -v -o website/oshiv/downloads/windows/intel/${OUT}_${VERSION}_windows_amd64 -ldflags="-X main.version=${VERSION}"
	GOOS=windows GOARCH=arm64 go build -v -o website/oshiv/downloads/windows/arm/${OUT}_${VERSION}_windows_arm64 -ldflags="-X main.version=${VERSION}"
	GOOS=linux GOARCH=amd64 go build -v -o website/oshiv/downloads/linux/intel/${OUT}_${VERSION}_linux_amd64 -ldflags="-X main.version=${VERSION}"
	GOOS=linux GOARCH=arm64 go build -v -o website/oshiv/downloads/linux/arm/${OUT}_${VERSION}_linux_arm64 -ldflags="-X main.version=${VERSION}"

zip:
	zip -j website/oshiv/downloads/mac/intel/${OUT}_${VERSION}_darwin_amd64.zip website/oshiv/downloads/mac/intel/${OUT}_${VERSION}_darwin_amd64
	zip -j website/oshiv/downloads/mac/arm/${OUT}_${VERSION}_darwin_arm64.zip website/oshiv/downloads/mac/arm/${OUT}_${VERSION}_darwin_arm64
	zip -j website/oshiv/downloads/windows/intel/${OUT}_${VERSION}_windows_amd64.zip website/oshiv/downloads/windows/intel/${OUT}_${VERSION}_windows_amd64
	zip -j website/oshiv/downloads/windows/arm/${OUT}_${VERSION}_windows_arm64.zip website/oshiv/downloads/windows/arm/${OUT}_${VERSION}_windows_arm64
	zip -j website/oshiv/downloads/linux/intel/${OUT}_${VERSION}_linux_amd64.zip website/oshiv/downloads/linux/intel/${OUT}_${VERSION}_linux_amd64
	zip -j website/oshiv/downloads/linux/arm/${OUT}_${VERSION}_linux_arm64.zip website/oshiv/downloads/linux/arm/${OUT}_${VERSION}_linux_arm64

html:
	cd website/oshiv; go run renderhtml.go ${VERSION} index.tmpl; cd ..
