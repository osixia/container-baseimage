NAME = osixia/light-baseimage
VERSION = 1.2.0

.PHONY: build build-nocache test tag-latest push push-latest release git-tag-version

build:
	docker buildx build --platform linux/arm/v7,linux/arm64/v8,linux/amd64 --load -t $(NAME):$(VERSION) -f image/Dockerfile image

build-nocache:
	docker buildx build --platform linux/arm/v7,linux/arm64/v8,linux/amd64 --load -t $(NAME):$(VERSION) -f image/Dockerfile --no-cache image

test:
	env NAME=$(NAME) VERSION=$(VERSION) bats test/test.bats

tag:
	docker tag $(NAME):$(VERSION) $(NAME):$(VERSION)

tag-latest:
	docker tag $(NAME):$(VERSION) $(NAME):latest

push:
	docker buildx build --platform linux/arm/v7,linux/arm64/v8,linux/amd64 --push -t $(NAME):$(VERSION) -f image/Dockerfile image

push-latest:
	docker buildx build --platform linux/arm/v7,linux/arm64/v8,linux/amd64 --push -t $(NAME):latest: -f image/Dockerfile image

release: build test tag-latest push push-latest

git-tag-version: release
	git tag -a v$(VERSION) -m "v$(VERSION)"
	git push origin v$(VERSION)
