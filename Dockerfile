# Note: this Dockerfile is actually used and extended by dagger to build all container images
# see build directory
ARG GOLANG_IMAGE="golang:1.25"
ARG BASE_IMAGE="debian:trixie-slim"

# step 1: build container-baseimage
FROM ${GOLANG_IMAGE} AS build

ARG IMAGE="osixia/baseimage:develop"

ARG VERSION="develop"
ARG CONTRIBUTORS="🐒✨🌴"

ARG GOARCH="amd64"

ENV GOOS="linux" \
    GOARCH="${GOARCH}" \
    CGO_ENABLED=0

RUN mkdir /build
WORKDIR /build

COPY . .

RUN go build \
    -ldflags="-w -s -X 'github.com/osixia/container-baseimage/config.Version=${VERSION}' -X 'github.com/osixia/container-baseimage/config.Contributors=${CONTRIBUTORS}' -X 'github.com/osixia/container-baseimage/config.ImageName=${IMAGE%%:*}' -X 'github.com/osixia/container-baseimage/config.ImageTag=${IMAGE##*:}'" \
    -o container \
    main.go

# step 2: create image
FROM ${BASE_IMAGE}

ARG BUILD_LOG_LEVEL="info"

COPY --from=build /build/container /usr/sbin/container
RUN container install --log-level ${BUILD_LOG_LEVEL}

ENV LANG="en_US.UTF-8" \
    LANGUAGE="en_US:en" \
    LC_ALL="en_US.UTF-8" \
    LC_CTYPE="en_US.UTF-8"

ENTRYPOINT ["/usr/sbin/container", "entrypoint"]
