#!/bin/sh -ex

## Add bash tools to /sbin
ln -s /container/tool/* /sbin/

mkdir -p /container/service
mkdir -p /container/environment /container/environment/startup
chmod 700 /container/environment/ /container/environment/startup

groupadd -g 8377 docker_env

# dpkg options
cp /container/file/dpkg_nodoc /etc/dpkg/dpkg.cfg.d/01_nodoc
cp /container/file/dpkg_nolocales /etc/dpkg/dpkg.cfg.d/01_nolocales

export LC_ALL=C
export DEBIAN_FRONTEND=noninteractive
minimal_apt_get_install='apt-get install -y --no-install-recommends'

## Prevent initramfs updates from trying to run grub and lilo.
## https://journal.paul.querna.org/articles/2013/10/15/docker-ubuntu-on-rackspace/
## http://bugs.debian.org/cgi-bin/bugreport.cgi?bug=594189
export INITRD=no
echo -n no > /container/environment/INITRD

## Enable Ubuntu Universe, Multiverse, and deb-src for main.
sed -i 's/^#\s*\(deb.*main restricted\)$/\1/g' /etc/apt/sources.list
sed -i 's/^#\s*\(deb.*universe\)$/\1/g' /etc/apt/sources.list
sed -i 's/^#\s*\(deb.*multiverse\)$/\1/g' /etc/apt/sources.list
apt-get update

## Fix some issues with APT packages.
## See https://github.com/dotcloud/docker/issues/1024
dpkg-divert --local --rename --add /sbin/initctl
ln -sf /bin/true /sbin/initctl

## Replace the 'ischroot' tool to make it always return true.
## Prevent initscripts updates from breaking /dev/shm.
## https://journal.paul.querna.org/articles/2013/10/15/docker-ubuntu-on-rackspace/
## https://bugs.launchpad.net/launchpad/+bug/974584
dpkg-divert --local --rename --add /usr/bin/ischroot
ln -sf /bin/true /usr/bin/ischroot

# apt-utils fix for Ubuntu 16.04
$minimal_apt_get_install apt-utils

## Install HTTPS support for APT.
$minimal_apt_get_install apt-transport-https ca-certificates

## Install add-apt-repository
$minimal_apt_get_install software-properties-common

## Upgrade all packages.
apt-get dist-upgrade -y --no-install-recommends -o Dpkg::Options::="--force-confold"

## Fix locale.
$minimal_apt_get_install language-pack-en
locale-gen en_US
update-locale LANG=en_US.UTF-8 LC_CTYPE=en_US.UTF-8
echo -n en_US.UTF-8 > /container/environment/LANG
echo -n en_US.UTF-8 > /container/environment/LANGUAGE
echo -n en_US.UTF-8 > /container/environment/LC_CTYPE

## Add python-yaml
$minimal_apt_get_install python3-yaml

apt-get clean
rm -rf /tmp/* /var/tmp/*
rm -rf /var/lib/apt/lists/*
rm -f /etc/dpkg/dpkg.cfg.d/02apt-speedup

# Remove useless files
rm -rf /container/file
rm -rf /container/build.sh /container/Dockerfile
