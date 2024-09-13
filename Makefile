OUT := oshiv
PKG := github.com/dnlloyd/oshiv
VERSION := $(shell git describe --always)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)
OS := $(shell uname -s | awk '{print tolower($0)}')
ARCH := $(shell uname -m)

build: vet staticcheck install-local

release: clean vet staticcheck compile zip html

clean:
	-@rm -fr website/downloads/mac/intel/${OUT}*
	-@rm -fr website/downloads/mac/arm/${OUT}*
	-@rm -fr website/downloads/windows/intel/${OUT}*
	-@rm -fr website/downloads/windows/arm/${OUT}*
	-@rm -fr website/downloads/linux/intel/${OUT}*
	-@rm -fr website/downloads/linux/arm/${OUT}*
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
	GOOS=darwin GOARCH=amd64 go build -v -o website/downloads/mac/intel/${OUT}_${VERSION}_darwin_amd64 -ldflags="-X main.version=${VERSION}"
	GOOS=darwin GOARCH=arm64 go build -v -o website/downloads/mac/arm/${OUT}_${VERSION}_darwin_arm64 -ldflags="-X main.version=${VERSION}"
	GOOS=windows GOARCH=amd64 go build -v -o website/downloads/windows/intel/${OUT}_${VERSION}_windows_amd64 -ldflags="-X main.version=${VERSION}"
	GOOS=windows GOARCH=arm64 go build -v -o website/downloads/windows/arm/${OUT}_${VERSION}_windows_arm64 -ldflags="-X main.version=${VERSION}"
	GOOS=linux GOARCH=amd64 go build -v -o website/downloads/linux/intel/${OUT}_${VERSION}_linux_amd64 -ldflags="-X main.version=${VERSION}"
	GOOS=linux GOARCH=arm64 go build -v -o website/downloads/linux/arm/${OUT}_${VERSION}_linux_arm64 -ldflags="-X main.version=${VERSION}"

zip:
	zip -j website/downloads/mac/intel/${OUT}_${VERSION}_darwin_amd64.zip website/downloads/mac/intel/${OUT}_${VERSION}_darwin_amd64
	zip -j website/downloads/mac/arm/${OUT}_${VERSION}_darwin_arm64.zip website/downloads/mac/arm/${OUT}_${VERSION}_darwin_arm64
	zip -j website/downloads/windows/intel/${OUT}_${VERSION}_windows_amd64.zip website/downloads/windows/intel/${OUT}_${VERSION}_windows_amd64
	zip -j website/downloads/windows/arm/${OUT}_${VERSION}_windows_arm64.zip website/downloads/windows/arm/${OUT}_${VERSION}_windows_arm64
	zip -j website/downloads/linux/intel/${OUT}_${VERSION}_linux_amd64.zip website/downloads/linux/intel/${OUT}_${VERSION}_linux_amd64
	zip -j website/downloads/linux/arm/${OUT}_${VERSION}_linux_arm64.zip website/downloads/linux/arm/${OUT}_${VERSION}_linux_arm64

html:
	cd website; go run renderhtml.go ${VERSION} index.tmpl; cd ..
