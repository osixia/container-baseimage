NAME = osixia/light-baseimage
VERSION = alpine-0.1.8

.PHONY: build build-nocache test tag-latest push push-latest release git-tag-version

build:
	docker build -f image/Dockerfile -t $(NAME):$(VERSION) --rm image

build-nocache:
	docker build -f image/Dockerfile -t $(NAME):$(VERSION) --no-cache --rm image

test:
	env NAME=$(NAME) VERSION=$(VERSION) bats test/test.bats

tag:
	docker tag $(NAME):$(VERSION) $(NAME):$(VERSION)

push:
	docker push $(NAME):$(VERSION)


release: build test tag-latest push push-latest

git-tag-version:
	git tag -a alpine-v$(VERSION:alpine-%=%) -m "v$(VERSION:alpine-%=%)"
	git push origin alpine-v$(VERSION:alpine-%=%)
