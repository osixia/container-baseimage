#!/bin/bash -e

# if container log level is trace:
# print commands and their arguments as they are executed
container logger level eq trace && set -x

# config
sed -i -e "s/expose_php = On/expose_php = Off/g" /etc/php/*/fpm/php.ini
sed -i -e "s/;cgi.fix_pathinfo=1/cgi.fix_pathinfo=0/g" /etc/php/*/fpm/php.ini
sed -i -e "s/;listen.owner = www-data/listen.owner = www-data/g" /etc/php/*/fpm/php.ini
sed -i -e "s/;listen.group = www-data/listen.group = www-data/g" /etc/php/*/fpm/php.ini
sed -i -e "s|listen = .*|listen = /run/container/php-fpm.sock|g" /etc/php/*/fpm/pool.d/www.conf

# replace default website with php-fpm service default website
cp -f /container/services/php-fpm/config/default /etc/nginx/sites-available/default

# create phpinfo.php
echo "<?php phpinfo(); " > /var/www/html/phpinfo.php
