#!/bin/bash -e

# install apache
LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends syslog-ng-core

mkdir -p /var/lib/syslog-ng
rm -f /etc/default/syslog-ng

touch /var/log/syslog
chmod u=rw,g=r,o= /var/log/syslog
rm -f /etc/syslog-ng/syslog-ng.conf
