#!/bin/bash -e
# this script is run during the image build

rm -rf /var/www/html/index.nginx-debian.html
echo "Hi!" > /var/www/html/index.html
