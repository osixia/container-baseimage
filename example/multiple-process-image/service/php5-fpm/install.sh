#!/bin/bash -e
# this script is run during the image build

# config
sed -i -e "s/expose_php = On/expose_php = Off/g" /etc/php5/php-fpm.conf
sed -i -e "s/;cgi.fix_pathinfo=1/cgi.fix_pathinfo=0/g" /etc/php5/php-fpm.conf

touch /etc/php5/fpm.d/www.conf
sed -i -e "s/;listen.owner = www-data/listen.owner = root/g" /etc/php5/fpm.d/www.conf
sed -i -e "s/;listen.group = www-data/listen.group = root/g" /etc/php5/fpm.d/www.conf

# replace default website with php5-fpm default website
cp -f /container/service/php5-fpm/config/nginx.conf /etc/nginx/nginx.conf

# create phpinfo.php
echo "<?php phpinfo(); " > /var/lib/nginx/html/phpinfo.php
