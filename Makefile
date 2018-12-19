# Build the project
all: clean build install

clean:
	rm -f cf-html5-apps-repo-cli-plugin

build:
	go build  

install:
	cf install-plugin -f cf-html5-apps-repo-cli-plugin	
