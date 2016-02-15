#!/bin/bash
log-helper level eq trace && set -x

exec tail -F -n 0 /var/log/syslog
