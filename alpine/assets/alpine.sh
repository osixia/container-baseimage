#!/bin/sh -e

# install required packages
container packages install --update --clean bash bash-completion libeatmydata

# add container bash completion
container completion bash > /usr/share/bash-completion/completions/container

# clean
rm -rf /tmp/* /var/tmp/*
