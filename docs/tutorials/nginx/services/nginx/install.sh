#!/bin/bash -e

# if container log level is trace:
# print commands and their arguments as they are executed
container logger level eq trace && set -x

# Remove default debian index file
rm -rf /var/www/html/index.nginx-debian.html

# Create default index file
echo "Hi!" > /var/www/html/index.html
