#!/bin/bash -e
# this script is run during the image build

# config
sed -i -e "s/expose_php = On/expose_php = Off/g" /etc/php5/php-fpm.conf
sed -i -e "s/;cgi.fix_pathinfo=1/cgi.fix_pathinfo=0/g" /etc/php5/php-fpm.conf
sed -i -e "s/;catch_workers_output\s*=\s*yes/catch_workers_output = yes/g" /etc/php5/php-fpm.conf
sed -i -e "s/listen = 127.0.0.1:9000/listen = \/var\/run\/php5-fpm.sock/g" /etc/php5/php-fpm.conf
sed -i -e "s/;listen.owner = nobody/listen.owner = nginx/g" /etc/php5/php-fpm.conf
sed -i -e "s/;listen.group = nobody/listen.group = www-data/g" /etc/php5/php-fpm.conf

# replace default website with php5-fpm default website
cp -f /container/service/php5-fpm/config/nginx.conf /etc/nginx/nginx.conf

# create phpinfo.php
echo "<?php phpinfo(); " > /var/lib/nginx/html/phpinfo.php

# fix dir and files permissions
chmod 755 /var/lib/nginx/ /var/lib/nginx/html/
chmod 644 /var/lib/nginx/html/phpinfo.php
