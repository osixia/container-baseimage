#!/bin/bash -e
log-helper level eq trace && set -x

FIRST_START_SETUP_DONE="/container/run/state/cfssl-first-start-setup-done"

# container first start setup
if [ ! -e "$FIRST_START_SETUP_DONE" ]; then

  chmod 700 ${SERVICE_DIR}/.cfssl/assets/tool/*
  ln -s ${SERVICE_DIR}/.cfssl/assets/tool/* /usr/sbin

  touch $FIRST_START_SETUP_DONE
fi
