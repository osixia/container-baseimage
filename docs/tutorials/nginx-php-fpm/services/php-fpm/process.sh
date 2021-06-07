#!/bin/bash -e

# if container log level is trace:
# print commands and their arguments as they are executed
container logger level eq trace && set -x

IP=$(hostname -i)
container logger info "Phpinfo available at: http://${IP}/phpinfo.php"

exec /usr/sbin/php-fpm* --nodaemonize "$@"
