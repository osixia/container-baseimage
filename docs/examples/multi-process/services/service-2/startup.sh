#!/bin/bash -e

# if container log level is trace:
# print commands and their arguments as they are executed
container log level eq trace && set -x

echo "service-2: Doing some container start setup ..."
echo "service-2: EXAMPLE_ENV_VAR=${EXAMPLE_ENV_VAR} ..."
