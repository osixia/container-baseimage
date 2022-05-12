#!/bin/sh -e
log-helper level eq trace && set -x

ln -sf "${CONTAINER_SERVICE_DIR}"/:ssl-tools/assets/tool/* /usr/sbin
