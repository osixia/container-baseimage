#!/bin/bash -e

FIRST_START_SETUP_DONE="/container/run/state/syslog-ng-first-start-setup-done"

# container first start setup
if [ ! -e "$FIRST_START_SETUP_DONE" ]; then

  ln -s ${SERVICE_DIR}/syslog-ng-core/assets/config/syslog_ng_default /etc/default/syslog-ng
  ln -s ${SERVICE_DIR}/syslog-ng-core/assets/config/syslog-ng.conf /etc/syslog-ng/syslog-ng.conf

  ## Install syslog to "docker logs" forwarder.
  mkdir /container/run/process/syslog-forwarder
  ln -s ${SERVICE_DIR}/syslog-ng-core/process-syslog-forwarder.sh /container/run/process/syslog-forwarder/run

  touch $FIRST_START_SETUP_DONE
fi
