ARG GOLANG_IMAGE="golang:1.25"
ARG BASE_IMAGE="debian:trixie-slim"

# step 1: build container binary
FROM ${GOLANG_IMAGE} AS build

ARG IMAGE="osixia/baseimage:develop"
ARG VERSION="develop"

ARG GOARCH="amd64"

ENV GOOS="linux" \
    GOARCH="${GOARCH}" \
    CGO_ENABLED=0

RUN mkdir /build
WORKDIR /build

COPY . .

RUN go build \
    -ldflags="-w -s -X 'github.com/osixia/container-baseimage/config.Version=${VERSION}' -X 'github.com/osixia/container-baseimage/config.ImageName=${IMAGE%%:*}' -X 'github.com/osixia/container-baseimage/config.ImageTag=${IMAGE##*:}'" \
    -o container \
    main.go

# step 2: create base image
FROM ${BASE_IMAGE} AS base

ARG BUILD_LOG_LEVEL="info"

COPY --from=build /build/container /usr/sbin/container
RUN container install --log-level ${BUILD_LOG_LEVEL}

ENV LANG="en_US.UTF-8" \
    LANGUAGE="en_US:en" \
    LC_ALL="en_US.UTF-8" \
    LC_CTYPE="en_US.UTF-8"

ENTRYPOINT ["/usr/sbin/container"]

# step 3: create nonroot image
FROM base AS nonroot

ARG NONROOT_GROUP_ID="65532"
ARG NONROOT_GROUP_NAME="nonroot"

ARG NONROOT_USER_ID="65532"
ARG NONROOT_USER_NAME="nonroot"

RUN container groups add ${NONROOT_GROUP_ID} ${NONROOT_GROUP_NAME} \
    && container users add ${NONROOT_USER_ID} ${NONROOT_USER_NAME} --group-id ${NONROOT_GROUP_NAME} --group-name ${NONROOT_GROUP_NAME}

USER ${NONROOT_USER_NAME}
