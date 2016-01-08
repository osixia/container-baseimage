#!/bin/sh
log-helper level is eq trace && set -x

exec tail -F -n 0 /var/log/syslog
