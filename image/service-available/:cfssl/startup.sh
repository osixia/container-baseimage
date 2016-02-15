#!/bin/sh -e
log-helper level eq trace && set -x

chmod 700 ${CONTAINER_SERVICE_DIR}/:cfssl/assets/tool/*
ln -sf ${CONTAINER_SERVICE_DIR}/:cfssl/assets/tool/* /usr/sbin
