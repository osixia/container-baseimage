#!/bin/bash -e

# set apt configuration (don't install recommended and suggested packages by default)
echo "APT::Install-Recommends \"false\";" > /etc/apt/apt.conf.d/99norecommends
echo "APT::Install-Suggests \"false\";" > /etc/apt/apt.conf.d/99nosuggests

# install required packages
container packages install --update --clean bash-completion locales eatmydata

# set locale
echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen
locale-gen en_US.UTF-8 2>&1 | container logger info
update-locale LANG=en_US.UTF-8 LC_CTYPE=en_US.UTF-8

# add container bash completion
container completion bash > /usr/share/bash-completion/completions/container
echo ". /etc/profile.d/bash_completion.sh" >> /root/.bashrc

# clean
rm -rf /tmp/* /var/tmp/*
