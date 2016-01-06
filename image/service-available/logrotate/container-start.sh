#!/bin/bash -e

FIRST_START_SETUP_DONE="/container/run/state/logrotate-first-start-setup-done"

# container first start setup
if [ ! -e "$FIRST_START_SETUP_DONE" ]; then

  ln -s ${SERVICE_DIR}/logrotate/assets/config/logrotate_syslogng /etc/logrotate.d/syslog-ng

  touch $FIRST_START_SETUP_DONE
fi
