#!/bin/bash -e
# this script is run during the image build

# config
sed -i -e "s/expose_php = On/expose_php = Off/g" /etc/php/7.0/fpm/php.ini
sed -i -e "s/;cgi.fix_pathinfo=1/cgi.fix_pathinfo=0/g" /etc/php/7.0/fpm/php.ini
sed -i -e "s/;listen.owner = www-data/listen.owner = www-data/g" /etc/php/7.0/fpm/php.ini
sed -i -e "s/;listen.group = www-data/listen.group = www-data/g" /etc/php/7.0/fpm/php.ini

# create php socket directory
mkdir -p /run/php

# replace default website with php service default website
cp -f /container/service/php/config/default /etc/nginx/sites-available/default

# create phpinfo.php
echo "<?php phpinfo(); " > /var/www/html/phpinfo.php
