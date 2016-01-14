#!/bin/bash -e
log-helper level eq trace && set -x

exec /usr/sbin/cron -f
