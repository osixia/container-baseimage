#!/bin/sh -e

# Install required packages.
apk add --update bash bash-completion libeatmydata

# Add container-baseimage bash completion.
container-baseimage completion bash > /usr/share/bash-completion/completions/container-baseimage

# Clean.
rm -rf /tmp/* /var/tmp/* /var/cache/apk/*
