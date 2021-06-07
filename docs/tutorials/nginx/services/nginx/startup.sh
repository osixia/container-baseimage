#!/bin/bash

set -euo pipefail

# if container log level is trace:
# print commands and their arguments as they are executed
container log level eq trace && set -x

FIRST_START_DONE="/run/container/nginx-first-start-done"

# container first start
if [ ! -e "${FIRST_START_DONE}" ]; then
  echo "${CUSTOM_MESSAGE}"  >> /var/www/html/index.html
  touch "${FIRST_START_DONE}"
fi
