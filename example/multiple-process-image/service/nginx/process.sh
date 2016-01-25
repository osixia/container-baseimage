#!/bin/bash -e
echo "The secret is: $FIRST_START_SETUP_ONLY_SECRET"
exec /usr/sbin/nginx -g "daemon off;"
