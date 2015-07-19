#!/bin/bash -e

# install apache
LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends syslog-ng-core

mkdir -p /var/lib/syslog-ng
cp /osixia/service-available/syslog-ng-core/assets/config/syslog_ng_default /etc/default/syslog-ng
touch /var/log/syslog
chmod u=rw,g=r,o= /var/log/syslog
cp /osixia/service-available/syslog-ng-core/assets/config/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf

## Install syslog to "docker logs" forwarder.
mkdir /etc/service/syslog-forwarder
ln -s /osixia/service-available/syslog-ng-core/daemon-syslog-forwarder.sh /etc/service/syslog-forwarder/run
