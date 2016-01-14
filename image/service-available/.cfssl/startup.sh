#!/bin/bash -e
log-helper level eq trace && set -x

FIRST_START_DONE="${CONTAINER_STATE_DIR}/cfssl-first-start-done"

# container first start
if [ ! -e "$FIRST_START_DONE" ]; then

  chmod 700 ${CONTAINER_SERVICE_DIR}/.cfssl/assets/tool/*
  ln -s ${CONTAINER_SERVICE_DIR}/.cfssl/assets/tool/* /usr/sbin

  touch $FIRST_START_DONE
fi
