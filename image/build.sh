#!/bin/sh -ex

## Add bash tools to /sbin
ln -s /container/tool/* /sbin/

mkdir -p /container/service
mkdir -p /container/environment /container/environment/startup
chmod 700 /container/environment/ /container/environment/startup

addgroup -g 8377 docker_env

# General config
export LC_ALL=C

## Prevent initramfs updates from trying to run grub and lilo.
## https://journal.paul.querna.org/articles/2013/10/15/docker-ubuntu-on-rackspace/
## http://bugs.debian.org/cgi-bin/bugreport.cgi?bug=594189
export INITRD=no
echo -n no > /container/environment/INITRD
echo -n C.UTF-8 > /container/environment/LANG
echo -n C.UTF-8 > /container/environment/LANGUAGE
echo -n C.UTF-8 > /container/environment/LC_CTYPE

## Install bash and python apt-utils.
apk add --update bash python3 py-yaml

rm -rf /var/cache/apk/*
rm -rf /tmp/* /var/tmp/*

# Remove useless files
rm -rf /container/build.sh /container/Dockerfile
