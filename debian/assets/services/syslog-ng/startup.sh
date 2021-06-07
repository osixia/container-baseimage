#!/bin/bash -e
container-logger level eq trace && set -x

# determine output mode on /dev/stdout because of the issue documented at https://github.com/phusion/baseimage-docker/issues/468
if [ -p /dev/stdout ]; then
    SYSLOG_OUTPUT_MODE_DEV_STDOUT="pipe"
else
    SYSLOG_OUTPUT_MODE_DEV_STDOUT="file"
fi

export SYSLOG_OUTPUT_MODE_DEV_STDOUT
envsubst-templates /container/services/syslog-ng/config /etc/syslog-ng
