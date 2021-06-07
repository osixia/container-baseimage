#!/bin/bash

set -euo pipefail

# if container log level is trace:
# print commands and their arguments as they are executed
container log level eq trace && set -x

IP=$(hostname -i)
container log info "Nginx is running at: http://${IP}"

exec /usr/sbin/nginx -g "daemon off;" "$@"
