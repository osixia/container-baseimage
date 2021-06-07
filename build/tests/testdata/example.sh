#!/bin/sh
set -eu

log=/tmp/baseimage-example.log
runner=
cleanup() {
  if [ -n "$runner" ]; then
    kill -TERM "$runner" 2>/dev/null || true
    wait "$runner" 2>/dev/null || true
  fi
  cat "$log"
}
trap cleanup EXIT

if [ "$1" = single-process ]; then
  # The single-process default must exit naturally without restarting.
  container --kill-all=false > "$log" 2>&1
else
  # The multi-process default must stay available for manual stop/start.
  container --kill-all=false > "$log" 2>&1 &
  runner=$!
  attempt=0
  until container processes check --quiet service-1 service-2 2>/dev/null; do
    attempt=$((attempt + 1)); [ "$attempt" -lt 100 ]
    kill -0 "$runner"
    sleep 0.1
  done
  container processes stop --timeout 10 service-1
  container processes check service-1 service-2
  container processes start --timeout 10 service-1
  kill -TERM "$runner"
  wait "$runner"
  runner=
  grep -Fq 'service-2: EXAMPLE_ENV_VAR=Hello :)' "$log"
  grep -Fq 'service-2: process ended' "$log"
fi
grep -Fq 'service-1: EXAMPLE_ENV_VAR=Hello :)' "$log"
grep -Fq 'service-1: Just going to sleep' "$log"
grep -Fq 'service-1: process ended' "$log"
[ ! -e /run/container/processes/service-1.pid ]

# Selection must exclude the other service's startup hook too.
container --exec service-1 --step startup --kill-all=false > /tmp/baseimage-selection.log 2>&1
grep -Fq 'service-1: EXAMPLE_ENV_VAR=Hello :)' /tmp/baseimage-selection.log
if grep -Fq 'service-2:' /tmp/baseimage-selection.log; then
  echo 'unselected service ran its startup hook' >&2
  exit 1
fi
