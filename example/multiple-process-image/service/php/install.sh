#!/bin/bash -e
# this script is run during the image build

# config
sed -i -e "s/;catch_workers_output\s*=\s*yes/catch_workers_output = yes/g" /etc/php7/php-fpm.d/www.conf
sed -i -e "s/listen = 127.0.0.1:9000/listen = \/run\/php\/php7.0-fpm.sock/g" /etc/php7/php-fpm.d/www.conf
sed -i -e "s/;listen.owner = nobody/listen.owner = nginx/g" /etc/php7/php-fpm.d/www.conf
sed -i -e "s/;listen.group = nobody/listen.group = www-data/g" /etc/php7/php-fpm.d/www.conf

# create php socket directory
mkdir -p /run/php

# replace default website with php service default website
cp -f /container/service/php/config/nginx.conf /etc/nginx/nginx.conf

# create phpinfo.php
echo "<?php phpinfo(); " > /var/lib/nginx/html/phpinfo.php

# fix dir and files permissions
chmod 755 /var/lib/nginx/ /var/lib/nginx/html/
chmod 644 /var/lib/nginx/html/phpinfo.php
