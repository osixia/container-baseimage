#!/bin/bash -e

set -o pipefail

# install required packages
container packages install --update --clean bash-completion locales eatmydata

# set locale
echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen
locale-gen en_US.UTF-8 2>&1 | container log info
update-locale LANG=en_US.UTF-8 LC_CTYPE=en_US.UTF-8 2>&1 | container log info

# add container bash completion
container completion bash > /usr/share/bash-completion/completions/container
echo ". /etc/profile.d/bash_completion.sh" >> /root/.bashrc

# clean
rm -frv /tmp/* /var/tmp/* 2>&1 | container log info
