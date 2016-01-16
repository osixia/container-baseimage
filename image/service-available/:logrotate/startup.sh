#!/bin/bash -e
log-helper level eq trace && set -x

FIRST_START_DONE="${CONTAINER_STATE_DIR}/logrotate-first-start-done"

# container first start
if [ ! -e "$FIRST_START_DONE" ]; then

  ln -s ${CONTAINER_SERVICE_DIR}/:logrotate/assets/config/logrotate_syslogng /etc/logrotate.d/syslog-ng

  touch $FIRST_START_DONE
fi
