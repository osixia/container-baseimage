#!/bin/sh
set -eu

case "${BASEIMAGE_TEST_MODE}" in
  exit) exit 7 ;;
  signal)
    trap 'exit 0' TERM INT
    touch /run/container/check-ready
    while :; do sleep 1; done
    ;;
  restart)
    echo attempt >> /run/container/check-attempts
    count=$(wc -l < /run/container/check-attempts)
    if [ "$count" -lt 2 ]; then exit 1; fi
    kill -TERM "$PPID"
    exec sleep 30
    ;;
  backoff)
    echo attempt >> /run/container/check-backoff-attempts
    if [ -f /run/container/check-stay-up ]; then exec sleep 30; fi
    exit 1
    ;;
esac
