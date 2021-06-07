#!/bin/bash -e

rm -f /etc/logrotate.conf /etc/logrotate.d/syslog-ng

ln -sf /container/services/logrotate/config/logrotate.conf /etc/logrotate.conf
ln -sf /container/services/logrotate/config/syslog-ng /etc/logrotate.d/syslog-ng

chmod 444 -R /container/services/logrotate/config/*
