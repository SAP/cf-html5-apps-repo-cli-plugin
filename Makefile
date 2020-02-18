# Build the project
all: clean build install

clean:
	rm -f cf-html5-apps-repo-cli-plugin

build:
	go build  

install:
	cf install-plugin -f cf-html5-apps-repo-cli-plugin	

release:
	rm -rf dist 
	mkdir dist
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -gcflags "all=-trimpath=${PWD}" -o dist/cf-html5-apps-repo-cli-plugin-darwin-amd64
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -gcflags "all=-trimpath=${PWD}" -o dist/cf-html5-apps-repo-cli-plugin-linux-amd64
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -gcflags "all=-trimpath=${PWD}" -o dist/cf-html5-apps-repo-cli-plugin-windows-amd64.exe 
