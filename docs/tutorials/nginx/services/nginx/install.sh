#!/bin/bash -e
# this script is run during the image build

# Remove default debian index file
rm -rf /var/www/html/index.nginx-debian.html

# Create default index file
echo "Hi!" > /var/www/html/index.html
