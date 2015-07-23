#!/bin/bash -e
if [ ! -d "/etc/ssl/certs/" ]; then
  mkdir -p /etc/ssl/certs/
fi

if [ ! -d "/etc/ssl/private/" ]; then
  mkdir -p /etc/ssl/private/
fi

if [ ! -e "/sbin/ssl-helper" ]; then
  # Add ssl-helper tool to sbin
  ln -s /container/service-available/ssl-helper/assets/tool/ssl-helper.sh /sbin/ssl-helper
fi
