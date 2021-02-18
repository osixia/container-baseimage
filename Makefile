NAME = osixia/light-baseimage
VERSION = 0.1.8

.PHONY: build build-nocache test tag-latest push push-latest release git-tag-version

build:
	docker build -f image/Dockerfile -t $(NAME):alpine-$(VERSION) --rm image

build-nocache:
	docker build -f image/Dockerfile -t $(NAME):alpine-$(VERSION) --no-cache --rm image

test:
	env NAME=$(NAME) VERSION=alpine-$(VERSION) bats test/test.bats

tag:
	docker tag $(NAME):alpine-$(VERSION) $(NAME):$(VERSION)

push:
	docker push $(NAME):alpine-$(VERSION)


release: build test tag-latest push push-latest

git-tag-version: release
	git tag -a alpine-v$(VERSION) -m "v$(VERSION)"
	git push origin alpine-v$(VERSION)
