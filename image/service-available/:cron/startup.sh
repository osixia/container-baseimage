#!/bin/bash -e
log-helper level eq trace && set -x

touch /etc/crontab /etc/periodic/15min/* /etc/periodic/hourly/* /etc/periodic/daily/* /etc/periodic/weekly/* /etc/periodic/monthly/*
