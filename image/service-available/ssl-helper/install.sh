#!/bin/bash -e
mkdir -p /etc/ssl/certs/ /etc/ssl/private/

# Add ssl-helper tool to sbin
ln -s /osixia/service-available/ssl-helper/tool/ssl-helper.sh /sbin/ssl-helper
