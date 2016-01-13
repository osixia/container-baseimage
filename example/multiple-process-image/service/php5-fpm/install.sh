#!/bin/bash -e
# this script is run during the image build

# config
sed -i --follow-symlinks -e "s/expose_php = On/expose_php = Off/g" /etc/php5/fpm/php.ini
sed -i --follow-symlinks -e "s/;cgi.fix_pathinfo=1/cgi.fix_pathinfo=0/g" /etc/php5/fpm/php.ini
sed -i --follow-symlinks -e "s/;listen.owner = www-data/listen.owner = www-data/g" /etc/php5/fpm/pool.d/www.conf
sed -i --follow-symlinks -e "s/;listen.group = www-data/listen.group = www-data/g" /etc/php5/fpm/pool.d/www.conf

# replace default website with php5-fpm default website
cp -f /container/service/php5-fpm/config/default /etc/nginx/sites-available/default

# create phpinfo.php
echo "<?php phpinfo(); " > /var/www/html/phpinfo.php
