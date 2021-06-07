#!/bin/bash -e
container-logger level eq trace && set -x

PIDFILE="/run/container/syslog-ng.pid"
SYSLOGNG_OPTS=""

[ -r /etc/syslog-ng/syslog-ng ] && . /etc/syslog-ng/syslog-ng

exec /usr/sbin/syslog-ng --pidfile "$PIDFILE" -F $SYSLOGNG_OPTS
