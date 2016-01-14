#!/bin/bash -e
log-helper level eq trace && set -x

touch /etc/crontab /etc/cron.d/* /etc/cron.daily/* /etc/cron.hourly/* /etc/cron.monthly/* /etc/cron.weekly/*
