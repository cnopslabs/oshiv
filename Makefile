OUT := oshiv
PKG := github.com/dnlloyd/oshiv
VERSION := $(shell git describe --always)
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/)

build: clean compile zip install-local

compile:
	GOOS=darwin GOARCH=amd64 go build -v -o downloads/mac/intel/${OUT}_${VERSION}_darwin_amd64 -ldflags="-X main.version=${VERSION}"
	GOOS=darwin GOARCH=arm64 go build -v -o downloads/mac/arm/${OUT}_${VERSION}_darwin_arm64 -ldflags="-X main.version=${VERSION}"
	GOOS=windows GOARCH=amd64 go build -v -o downloads/windows/intel/${OUT}_${VERSION}_windows_amd64 -ldflags="-X main.version=${VERSION}"
	GOOS=windows GOARCH=arm64 go build -v -o downloads/windows/arm/${OUT}_${VERSION}_windows_arm64 -ldflags="-X main.version=${VERSION}"
	GOOS=linux GOARCH=amd64 go build -v -o downloads/linux/intel/${OUT}_${VERSION}_linux_amd64 -ldflags="-X main.version=${VERSION}"
	GOOS=linux GOARCH=arm64 go build -v -o downloads/linux/arm/${OUT}_${VERSION}_linux_arm64 -ldflags="-X main.version=${VERSION}"

install-local:
	GOOS=darwin GOARCH=arm64 go build -v -ldflags="-X main.version=${VERSION}"
	go install -v -ldflags="-X main.version=${VERSION}"

zip:
	zip -j downloads/mac/intel/${OUT}_${VERSION}_darwin_amd64.zip downloads/mac/intel/${OUT}_${VERSION}_darwin_amd64
	zip -j downloads/mac/arm/${OUT}_${VERSION}_darwin_arm64.zip downloads/mac/arm/${OUT}_${VERSION}_darwin_arm64
	zip -j downloads/windows/intel/${OUT}_${VERSION}_windows_amd64.zip downloads/windows/intel/${OUT}_${VERSION}_windows_amd64
	zip -j downloads/windows/arm/${OUT}_${VERSION}_windows_arm64.zip downloads/windows/arm/${OUT}_${VERSION}_windows_arm64
	zip -j downloads/linux/intel/${OUT}_${VERSION}_linux_amd64.zip downloads/linux/intel/${OUT}_${VERSION}_linux_amd64
	zip -j downloads/linux/arm/${OUT}_${VERSION}_linux_arm64.zip downloads/linux/arm/${OUT}_${VERSION}_linux_arm64

# test:
# 	@go test -short ${PKG_LIST}

# vet:
# 	@go vet ${PKG_LIST}

# lint:
# 	@for file in ${GO_FILES} ;  do \
# 		golint $$file ; \
# 	done

# static: vet lint
# 	go build -i -v -o ${OUT}-v${VERSION} -tags netgo -ldflags="-extldflags \"-static\" -w -s -X main.version=${VERSION}" ${PKG}

# run: compile
# 	./${OUT}

clean:
	-@rm -fr downloads/mac/intel/${OUT}*
	-@rm -fr downloads/mac/arm/${OUT}*
	-@rm -fr downloads/windows/intel/${OUT}*
	-@rm -fr downloads/windows/arm/${OUT}*
	-@rm -fr downloads/linux/intel/${OUT}*
	-@rm -fr downloads/linux/arm/${OUT}*

# .PHONY: run server static vet lint
