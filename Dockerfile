ARG UBUNTU_VERSION=22.04

ARG GOMPLATE_VERSION=v3.11.2-slim

ARG GOLANG_VERSION=1.18

FROM hairyhenderson/gomplate:$GOMPLATE_VERSION as gomplate

FROM golang:$GOLANG_VERSION as builder

RUN mkdir -p /opt/bin

# compile snip
ENV GOFLAGS=-mod=vendor
WORKDIR /snip
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o=/opt/bin/snip .

# bash completion
RUN mkdir -p /etc/bash_completion.d && \
  printf "#!/bin/sh\n. <(snip completion)">/etc/bash_completion.d/snip && \
  chmod +x /etc/bash_completion.d/snip

# final image
# FROM scratch

FROM ubuntu:$UBUNTU_VERSION

RUN apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install -yq --no-install-recommends \
    curl \
    ca-certificates \
    wget \
    git \
  && rm -rf /var/lib/apt/lists/*

ARG KUBECTL_VERSION=v1.25.0
ENV KUBECTL_VERSION=$KUBECTL_VERSION
RUN curl -sL https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl > /usr/local/bin/kubectl \
  && chmod +x /usr/local/bin/kubectl

ARG HELM_VERSION=v3.9.3
ENV HELM_VERSION=$HELM_VERSION
RUN curl -sL https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz | tar xz -C /tmp/ \
  && mv /tmp/linux-amd64/helm /usr/local/bin/helm \
  && chmod +x /usr/local/bin/helm \
  && rm -r /tmp/linux-amd64

ARG NODE_VERSION=18
ENV NODE_VERSION=$NODE_VERSION
RUN wget -qO- https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
  && apt-get install nodejs \
  && npm install -g yarn \
  && rm -rf /var/lib/apt/lists/*

ARG KAPP_VERSION=v0.52.0
ENV KAPP_VERSION=$KAPP_VERSION
RUN curl -sL https://github.com/vmware-tanzu/carvel-kapp/releases/download/${KAPP_VERSION}/kapp-linux-amd64 > /tmp/kapp \
  && mv /tmp/kapp /usr/local/bin/kapp \
  && chmod +x /usr/local/bin/kapp

ARG ROLLOUT_STATUS_VERSION=v1.9.0
ENV ROLLOUT_STATUS_VERSION=$ROLLOUT_STATUS_VERSION
RUN curl -sL https://github.com/SocialGouv/rollout-status/releases/download/${ROLLOUT_STATUS_VERSION}/rollout-status-${ROLLOUT_STATUS_VERSION}-linux-amd64 > /tmp/rollout-status \
  && mv /tmp/rollout-status /usr/local/bin/rollout-status \
  && chmod +x /usr/local/bin/rollout-status

RUN groupadd -g 1001 ubuntu && useradd -rm -d /home/ubuntu -s /bin/bash -g ubuntu -G sudo -u 1001 ubuntu
RUN mkdir -p /opt && chown 1001:1001 /opt

COPY --from=gomplate        /gomplate                       /usr/local/bin/gomplate
COPY --from=builder         /opt/bin/                       /usr/local/bin/
COPY --from=builder         /etc/bash_completion.d/snip      /etc/bash_completion.d/snip

ENTRYPOINT ["/usr/local/bin/snip"]
CMD ["help"]