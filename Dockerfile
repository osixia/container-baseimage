ARG GOLANG_IMAGE_NAME="golang"
ARG GOLANG_IMAGE_TAG="1.21"

ARG ROOT_IMAGE_NAME="debian"
ARG ROOT_IMAGE_TAG="bookworm-slim"

# step 1: build container-baseimage
FROM ${GOLANG_IMAGE_NAME}:${GOLANG_IMAGE_TAG} AS build

ARG BUILD_VERSION="develop"
ARG BUILD_CONTRIBUTORS="🐒✨🌴"

ARG BUILD_IMAGE_NAME="osixia/baseimage"
ARG BUILD_IMAGE_TAG="develop"

ARG GOARCH="amd64"

ENV GOOS="linux" \
    GOARCH="${GOARCH}" \
    CGO_ENABLED=0

RUN mkdir /build
WORKDIR /build

COPY . .

RUN go build \
    -ldflags="-w -s -X 'github.com/osixia/container-baseimage/config.BuildVersion=${BUILD_VERSION}' -X 'github.com/osixia/container-baseimage/config.BuildContributors=${BUILD_CONTRIBUTORS}' -X 'github.com/osixia/container-baseimage/config.BuildImageName=${BUILD_IMAGE_NAME}' -X 'github.com/osixia/container-baseimage/config.BuildImageTag=${BUILD_IMAGE_TAG}'" \
    -o container-baseimage \
    main.go

# step 2: create image
FROM ${ROOT_IMAGE_NAME}:${ROOT_IMAGE_TAG}

ARG BUILD_LOG_LEVEL="info"

COPY --from=build /build/container-baseimage /usr/sbin/container-baseimage
RUN container-baseimage install --log-level ${BUILD_LOG_LEVEL}

ENV LANG="en_US.UTF-8" \
    LANGUAGE="en_US:en" \
    LC_ALL="en_US.UTF-8" \
    LC_CTYPE="en_US.UTF-8"

ENTRYPOINT ["/usr/sbin/container-baseimage", "entrypoint"]
