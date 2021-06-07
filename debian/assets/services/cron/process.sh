#!/bin/bash -e
container-logger level eq trace && set -x

exec /usr/sbin/cron -f
