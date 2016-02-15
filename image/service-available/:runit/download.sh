#!/bin/sh -e

# download runit from apt-get
apk add runit --update-cache --repository http://dl-3.alpinelinux.org/alpine/edge/testing/ --allow-untrusted
