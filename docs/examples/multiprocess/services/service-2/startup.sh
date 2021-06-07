#!/bin/bash -e

# if container log level is trace:
# print commands and their arguments as they are executed
container-logger level eq trace && set -x

FIRST_START_DONE="/run/container/service-2-first-start-done"

if [ ! -e "${FIRST_START_DONE}" ]; then
    echo "service-2: Doing some container first start setup ..."

    touch "${FIRST_START_DONE}"
fi

echo "service-2: Doing some others container start setup ..."
echo "service-2: EXAMPLE_ENV_VAR=${EXAMPLE_ENV_VAR} ..."
