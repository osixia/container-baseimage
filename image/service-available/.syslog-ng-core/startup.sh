#!/bin/bash -e
log-helper level eq trace && set -x

FIRST_START_DONE="${CONTAINER_STATE_DIR}/syslog-ng-first-start-done"

# container first start
if [ ! -e "$FIRST_START_DONE" ]; then

  ln -s ${CONTAINER_SERVICE_DIR}/.syslog-ng-core/assets/config/syslog_ng_default /etc/default/syslog-ng
  ln -s ${CONTAINER_SERVICE_DIR}/.syslog-ng-core/assets/config/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf

  ## Install syslog to "docker logs" forwarder.
  mkdir /container/run/process/syslog-forwarder
  ln -s ${CONTAINER_SERVICE_DIR}/.syslog-ng-core/process-syslog-forwarder.sh /container/run/process/syslog-forwarder/run

  touch $FIRST_START_DONE
fi
