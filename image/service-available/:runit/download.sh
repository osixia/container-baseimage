#!/bin/sh -e

# download runit from apk
echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories
apk add --update runit
