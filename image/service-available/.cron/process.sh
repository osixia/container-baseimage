#!/bin/bash -e
log-helper level is eq trace && set -x

exec /usr/sbin/cron -f
