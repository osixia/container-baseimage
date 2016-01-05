#!/bin/bash -e

# install apache
LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends syslog-ng-core

mkdir -p /var/lib/syslog-ng
rm -f /etc/default/syslog-ng
ln -s /container/service-available/syslog-ng-core/assets/config/syslog_ng_default /etc/default/syslog-ng
touch /var/log/syslog
chmod u=rw,g=r,o= /var/log/syslog
rm -f /etc/syslog-ng/syslog-ng.conf
ln -s /container/service-available/syslog-ng-core/assets/config/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf

## Install syslog to "docker logs" forwarder.
mkdir /container/run/service/syslog-forwarder
ln -s /container/service-available/syslog-ng-core/daemon-syslog-forwarder.sh /container/run/service/syslog-forwarder/run
