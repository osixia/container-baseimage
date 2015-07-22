#!/bin/bash -e

# install php
LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends php5-fpm

# config
sed -i -e "s/expose_php = On/expose_php = Off/g" /etc/php5/fpm/php.ini
sed -i -e "s/;cgi.fix_pathinfo=1/cgi.fix_pathinfo=0/g" /etc/php5/fpm/php.ini
sed -i -e "s/;listen.owner = www-data/listen.owner = www-data/g" /etc/php5/fpm/pool.d/www.conf
sed -i -e "s/;listen.group = www-data/listen.group = www-data/g" /etc/php5/fpm/pool.d/www.conf

# nginx is installed
if [ -d "/etc/nginx" ]; then

  mkdir /etc/nginx/common
  ln -s /osixia/service/php5-fpm/assets/config/nginx/php5-fpm.conf /etc/nginx/common/php5-fpm.conf

fi

# apache2 is installed
if [ -d "/etc/apache2" ]; then

  echo "deb http://http.debian.net/debian/ jessie main contrib non-free" >> /etc/apt/sources.list

  apt-get update && LC_ALL=C DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends libapache2-mod-fastcgi

  a2enmod fastcgi actions alias

  ln -s /osixia/service/php5-fpm/assets/config/apache2/php5-fpm.conf /etc/apache2/conf-available/php5-fpm.conf
  a2enconf php5-fpm

fi
