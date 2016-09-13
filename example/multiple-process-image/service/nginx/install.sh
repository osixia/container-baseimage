#!/bin/bash -e
# this script is run during the image build

mkdir -p /run/nginx

rm -rf /var/lib/nginx/html/index.html
echo "Hi!" > /var/lib/nginx/html/index.html
