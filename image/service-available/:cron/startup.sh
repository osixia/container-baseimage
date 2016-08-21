#!/bin/bash -e
log-helper level eq trace && set -x

touch /etc/crontab /etc/cron.d /etc/cron.daily /etc/cron.hourly /etc/cron.monthly /etc/cron.weekly

find /etc/cron.d/ -exec touch {} \;
find /etc/cron.daily/ -exec touch {} \;
find /etc/cron.hourly/ -exec touch {} \;
find /etc/cron.monthly/ -exec touch {} \;
find /etc/cron.weekly/ -exec touch {} \;
