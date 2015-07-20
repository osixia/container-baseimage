#!/bin/bash -ex

## Install osixia tools
mkdir -p /etc/service

ln -s /osixia/tool/install-multiple-process-stack /sbin/install-multiple-process-stack
ln -s /osixia/tool/install-service /sbin/install-service
ln -s /osixia/tool/install-service-available /sbin/install-service-available
ln -s /osixia/tool/add-host /sbin/add-host

ln -s /osixia/tool/py_tool/my_init /sbin/my_init
ln -s /etc/service
mkdir -p /etc/my_init.d
mkdir -p /etc/container_environment
touch /etc/container_environment.sh
chmod 700 /etc/container_environment

groupadd -g 8377 docker_env
chown :docker_env /etc/container_environment.sh
chmod 640 /etc/container_environment.sh

ln -s /osixia/tool/py_tool/setuser /sbin/setuser

# dpkg
cp /osixia/file/dpkg_nodoc /etc/dpkg/dpkg.cfg.d/01_nodoc
cp /osixia/file/dpkg_nolocales /etc/dpkg/dpkg.cfg.d/01_nolocales

# Remove file directory
rm -rf /osixia/file

# General config
export LC_ALL=C
export DEBIAN_FRONTEND=noninteractive
minimal_apt_get_install='apt-get install -y --no-install-recommends'

## Temporarily disable dpkg fsync to make building faster.
if [[ ! -e /etc/dpkg/dpkg.cfg.d/docker-apt-speedup ]]; then
	echo force-unsafe-io > /etc/dpkg/dpkg.cfg.d/docker-apt-speedup
fi

## Prevent initramfs updates from trying to run grub and lilo.
## https://journal.paul.querna.org/articles/2013/10/15/docker-ubuntu-on-rackspace/
## http://bugs.debian.org/cgi-bin/bugreport.cgi?bug=594189
export INITRD=no
mkdir -p /etc/container_environment
echo -n no > /etc/container_environment/INITRD

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

## Install apt-utils.
$minimal_apt_get_install apt-utils

## Upgrade all packages.
apt-get dist-upgrade -y --no-install-recommends

apt-get clean
rm -rf /osixia/build.sh /osixia/Dockerfile
rm -rf /tmp/* /var/tmp/*
rm -rf /var/lib/apt/lists/*
rm -f /etc/dpkg/dpkg.cfg.d/02apt-speedup
