#!/bin/bash -e

# if container log level is trace:
# print commands and their arguments as they are executed
container log level eq trace && set -x

# remove default debian index file
rm -rf /var/www/html/index.nginx-debian.html

# create default index file
echo "Hi!" > /var/www/html/index.html
