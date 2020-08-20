PROJECT_NAME := "snip"
PKG := "gitlab.com/youtopia.earth/ops/$(PROJECT_NAME)"

all: vendor fmt build targz
install: install-bin autocomplete
buildstall: build install

docker:
	docker build -t registry.gitlab.com/youtopia.earth/ops/snip:$${SNIP_TAG:-master} .

vendor:
	go mod vendor

fmt:
	# gofmt -w .
	gofmt -w -l `find . -type f -name '*.go'| grep -v "/vendor/"`

build:
	CGO_ENABLED=0 GOOS=linux go build -o snip -v $(PKG) .

targz:
	tar -cvzf $(PROJECT_NAME).tar.gz $(PROJECT_NAME)

install-bin:
	sudo cp -f snip /usr/local/bin/snip

autocomplete:
	[ -f ~/.bashrc ] || touch ~/.bashrc
	grep -xq ". <(snip completion)" ~/.bashrc || printf "\n. <(snip completion)\n" >> ~/.bashrc


help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
