#!/bin/bash -e

# if container log level is trace:
# print commands and their arguments as they are executed
container-logger level eq trace && set -x

echo "service-1: process ended ..."
