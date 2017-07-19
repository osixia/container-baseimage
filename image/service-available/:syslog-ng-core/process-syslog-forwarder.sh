#!/bin/sh -e
log-helper level eq trace && set -x

exec tail -F -n 0 /var/log/syslog > /proc/1/fd/1
