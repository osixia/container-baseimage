#!/bin/bash -e
container-logger level eq trace && set -x

ln -sf /container/services/logrotate/assets/config/logrotate.conf /etc/logrotate.conf
ln -sf /container/services/logrotate/assets/config/logrotate_syslogng /etc/logrotate.d/syslog-ng

chmod 444 -R /container/services/logrotate/assets/config/*
