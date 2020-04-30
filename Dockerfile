ARG DOCKER_VERSION=19
ARG GOMPLATE_VERSION=v3.6.0-slim
ARG GOLANG_VERSION=1

FROM docker:$DOCKER_VERSION AS docker
FROM hairyhenderson/gomplate:$GOMPLATE_VERSION as gomplate
FROM golang:$GOLANG_VERSION as builder

RUN mkdir /opt/bin

# compile snip
ENV GOFLAGS=-mod=vendor
WORKDIR /snip
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o=/opt/bin/snip .

# bash completion
RUN mkdir -p /etc/bash_completion.d && \
  printf "#!/bin/sh\n. <(snip completion)">/etc/bash_completion.d/snip && \
  chmod +x /etc/bash_completion.d/snip

# create clean image
FROM scratch
COPY --from=docker          /usr/local/bin/docker           /usr/local/bin/docker
COPY --from=gomplate        /gomplate                       /usr/local/bin/gomplate
COPY --from=builder         /opt/bin/                       /usr/local/bin/
COPY --from=builder         /etc/bash_completion.d/snip      /etc/bash_completion.d/snip

ENTRYPOINT ["/usr/local/bin/snip"]
CMD ["help"]