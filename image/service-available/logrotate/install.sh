#!/bin/bash -e

# install logrotate
LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends logrotate
cp /osixia/service/logrotate/assets/config/logrotate_syslogng /etc/logrotate.d/syslog-ng
