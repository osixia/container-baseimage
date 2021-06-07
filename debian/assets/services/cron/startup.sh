#!/bin/bash -e
container-logger level eq trace && set -x

# prevent NUMBER OF HARD LINKS > 1 error
# https://github.com/phusion/baseimage-docker/issues/198
for dir in /etc/crontab /etc/cron.d /etc/cron.daily /etc/cron.hourly /etc/cron.monthly /etc/cron.weekly; do 
    find ${dir} -exec touch {} +
done
