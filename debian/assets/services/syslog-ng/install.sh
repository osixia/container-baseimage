#!/bin/bash -e

mkdir -p /var/lib/syslog-ng
rm -f /etc/default/syslog-ng /etc/syslog-ng/syslog-ng.conf

ln -sf /container/services/syslog-ng/config/syslog-ng /etc/default/syslog-ng

touch /var/log/syslog
chmod 640 /var/log/syslog

# If /var/log is writable by another user logrotate will fail
/bin/chown root:root /var/log
/bin/chmod 0755 /var/log
