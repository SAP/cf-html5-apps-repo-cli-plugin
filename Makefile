NAME=cf-html5-apps-repo-cli-plugin
VERSION=1.4.8

# Build the project
all: clean build install

clean:
	rm -f ${NAME}

build:
	GO111MODULE=off go build -ldflags="-s -w" -gcflags "all=-trimpath=${HOME}" 

install:
	cf install-plugin -f ${NAME}

release:
	rm -rf dist 
	mkdir dist
	GO111MODULE=off GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -gcflags "all=-trimpath=${HOME}" -o dist/${NAME}-darwin-amd64
	GO111MODULE=off GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -gcflags "all=-trimpath=${HOME}" -o dist/${NAME}-darwin-arm64
	GO111MODULE=off GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -gcflags "all=-trimpath=${HOME}" -o dist/${NAME}-linux-amd64
	GO111MODULE=off GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -gcflags "all=-trimpath=${HOME}" -o dist/${NAME}-linux-arm64
	GO111MODULE=off GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -gcflags "all=-trimpath=${HOME}" -o dist/${NAME}-windows-amd64.exe
	@echo "===================="
	@echo "- authors:"
	@echo "  - contact: micellius@gmail.com"
	@echo "    homepage: https://github.com/micellius"
	@echo "    name: micellius"
	@echo "  binaries:"
	@echo "  - checksum: $(shell shasum -a 1 dist/${NAME}-darwin-amd64 | awk '{print $$1}')"
	@echo "    platform: osx"
	@echo "    url: https://github.com/SAP/${NAME}/releases/download/v${VERSION}/${NAME}-darwin-amd64"
	@echo "  - checksum: $(shell shasum -a 1 dist/${NAME}-linux-amd64 | awk '{print $$1}')"
	@echo "    platform: linux64"
	@echo "    url: https://github.com/SAP/${NAME}/releases/download/v${VERSION}/${NAME}-linux-amd64"
	@echo "  - checksum: $(shell shasum -a 1 dist/${NAME}-windows-amd64.exe | awk '{print $$1}')"
	@echo "    platform: win64"
	@echo "    url: https://github.com/SAP/${NAME}/releases/download/v${VERSION}/${NAME}-windows-amd64.exe"
	@echo "  company: SAP"
	@echo "  created: 2019-02-05T12:00:00Z"
	@echo "  description: CLI client for SAP Cloud Platform HTML5 Applications Repository service"
	@echo "  homepage: https://sap.github.io/${NAME}"
	@echo "  name: html5-plugin"
	@echo "  updated: $(shell date -u +'%FT%TZ')"
	@echo "  version: ${VERSION}"
	@echo "===================="
