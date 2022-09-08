buildstall: build install autocomplete

build:
	CGO_ENABLED=0 GOOS=linux go build -a -mod vendor -installsuffix cgo -o snip .

install:
	sudo cp -f snip /usr/local/bin/snip

autocomplete:
	echo "source <(snip completion)" | sudo tee /etc/bash_completion.d/snip.sh

update:
	go mod tidy
	go mod vendor