#!/bin/sh -e
log-helper level eq trace && set -x

PIDFILE="/var/run/syslog-ng.pid"
SYSLOGNG_OPTS=""

[ -r /etc/syslog-ng/syslog-ng ] && . /etc/syslog-ng/syslog-ng

exec /usr/sbin/syslog-ng --pidfile "$PIDFILE" -F $SYSLOGNG_OPTS
