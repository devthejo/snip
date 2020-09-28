PROGNAME=snip
OUTPUT=${PROGNAME}_${CI_COMMIT_TAG}_${GOOS}_${GOARCH}
PKG=gitlab.com/ytopia/ops/$(PROGNAME)

all:
	go build
cross:
	go build -o ${OUTPUT}
	sha256sum ${OUTPUT} >${OUTPUT}.sha256

buildstall: build install autocomplete

build:
	go build -o snip -v $(PKG) .
install:
	sudo cp -f snip /usr/local/bin/snip
autocomplete:
	echo "source <(snip completion)" | sudo tee /etc/bash_completion.d/snip.sh
