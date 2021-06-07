#!/bin/bash -e

# if container log level is trace:
# print commands and their arguments as they are executed
container log level eq trace && set -x

SLEEP=$(shuf -i 3-15 -n 1)

echo "service-2: Just going to sleep for ${SLEEP} seconds ..."
exec sleep "${SLEEP}"
