buildstall: build install autocomplete

VERSION_TAG := $(shell git describe --tags $(git rev-list --tags --max-count=1))

build: update
	CGO_ENABLED=0 GOOS=linux go build -mod vendor -installsuffix cgo -ldflags="-X 'main.Version=${VERSION_TAG}'" -o snip .

install:
	sudo cp -f snip /usr/local/bin/snip

autocomplete:
	echo "source <(snip completion)" | sudo tee /etc/bash_completion.d/snip.sh

update:
	go mod tidy
	go mod vendor