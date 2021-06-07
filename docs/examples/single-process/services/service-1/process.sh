#!/bin/bash -e

# if container log level is trace:
# print commands and their arguments as they are executed
container-logger level eq trace && set -x

SLEEP=$(shuf -i 3-45 -n 1)

echo "service-1: Just going to sleep for ${SLEEP} seconds ..."
exec sleep "${SLEEP}"
