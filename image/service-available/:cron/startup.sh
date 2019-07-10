#!/bin/sh -e
log-helper level eq trace && set -x

# prevent NUMBER OF HARD LINKS > 1 error
# https://github.com/phusion/baseimage-docker/issues/198
touch /etc/crontabs /etc/periodic/15min /etc/periodic/hourly /etc/periodic/daily /etc/periodic/weekly /etc/periodic/monthly

find /etc/crontabs/ -exec touch {} \;
find /etc/periodic/15min/ -exec touch {} \;
find /etc/periodic/hourly/ -exec touch {} \;
find /etc/periodic/daily/ -exec touch {} \;
find /etc/periodic/weekly/ -exec touch {} \;
find /etc/periodic/monthly/ -exec touch {} \;
