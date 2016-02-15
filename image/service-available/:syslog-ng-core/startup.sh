#!/bin/bash -e
log-helper level eq trace && set -x

ln -sf ${CONTAINER_SERVICE_DIR}/:syslog-ng-core/assets/config/syslog_ng_default /etc/default/syslog-ng
ln -sf ${CONTAINER_SERVICE_DIR}/:syslog-ng-core/assets/config/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf

## Install syslog to "docker logs" forwarder.
[ -d /container/run/process/:syslog-forwarder ] || mkdir -p /container/run/process/:syslog-forwarder
ln -sf ${CONTAINER_SERVICE_DIR}/:syslog-ng-core/process-syslog-forwarder.sh /container/run/process/:syslog-forwarder/run
