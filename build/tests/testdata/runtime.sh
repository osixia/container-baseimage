#!/bin/sh
set -eu

runner=
cleanup() {
  if [ -n "$runner" ]; then
    kill -TERM "$runner" 2>/dev/null || true
    wait "$runner" 2>/dev/null || true
  fi
}
trap cleanup EXIT

wait_ready() {
  attempt=0
  while [ ! -f /run/container/check-ready ]; do
    attempt=$((attempt + 1))
    [ "$attempt" -lt 100 ] || { echo 'service did not become ready' >&2; exit 1; }
    kill -0 "$runner"
    sleep 0.1
  done
}

status=0
export BASEIMAGE_TEST_MODE=exit
container --skip-env-files --kill-all=false --exec runtime-check || status=$?
[ "$status" -eq 7 ]
[ -f /run/container/check-startup ]
[ -f /run/container/check-finish ]
rm /run/container/check-startup /run/container/check-finish

export BASEIMAGE_TEST_MODE=signal
container --skip-env-files --kill-all=false --exec runtime-check &
runner=$!
wait_ready
kill -TERM "$runner"
wait "$runner"
runner=
[ -f /run/container/check-finish ]
rm /run/container/check-finish /run/container/check-ready

export BASEIMAGE_TEST_MODE=restart
container --skip-env-files --kill-all=false --restart=true --exec runtime-check
[ "$(wc -l < /run/container/check-attempts)" -eq 2 ]
[ -f /run/container/check-finish ]

# Stop during an automatic restart delay: keep the process discoverable, then
# honor the stop request without another launch, and allow a later start.
export BASEIMAGE_TEST_MODE=backoff
container --skip-env-files --kill-all=false --restart=true --exec runtime-check &
runner=$!
attempt=0
while [ ! -f /run/container/processes/runtime-check.starting ] ||
      [ ! -f /run/container/check-backoff-attempts ] ||
      [ "$(wc -l < /run/container/check-backoff-attempts)" -lt 5 ]; do
  attempt=$((attempt + 1)); [ "$attempt" -lt 500 ]
  kill -0 "$runner"
  sleep 0.01
done
container processes status | grep -F 'runtime-check - Up:false, Starting:true'
status=0
container processes check runtime-check --quiet 2>/dev/null || status=$?
[ "$status" -eq 1 ]
count=$(wc -l < /run/container/check-backoff-attempts)
container processes stop --timeout 10 runtime-check
attempt=0
while [ -f /run/container/processes/runtime-check.starting ]; do
  attempt=$((attempt + 1)); [ "$attempt" -lt 100 ]; sleep 0.1
done
[ "$(wc -l < /run/container/check-backoff-attempts)" -eq "$count" ]
container processes check runtime-check
touch /run/container/check-stay-up
container processes start --timeout 10 runtime-check
kill -TERM "$runner"
wait "$runner"
runner=

# Closing Bash must end the process step even when services restart by default.
# A FIFO keeps its stdin open until the service has really started.
for expected in 0 9; do
  rm -f /run/container/check-ready /run/container/check-finish
  fifo="/run/container/check-stdin-$expected"
  mkfifo "$fifo"
  export BASEIMAGE_TEST_MODE=signal
  container --skip-env-files --kill-all=false --restart=true --exec runtime-check --bash < "$fifo" &
  runner=$!
  exec 3> "$fifo"
  wait_ready
  printf 'exit %s\n' "$expected" >&3
  exec 3>&-
  status=0
  wait "$runner" || status=$?
  runner=
  [ "$status" -eq "$expected" ]
  [ -f /run/container/check-finish ]
  [ ! -e /run/container/processes/runtime-check.pid ]
  rm "$fifo"
done

# No services and no command: Bash starts without --bash.
status=0
printf 'exit 11\n' | container --skip-env-files --skip-all --kill-all=false || status=$?
[ "$status" -eq 11 ]

# --skip-all also bypasses startup and finish; --skip-process --bash keeps them.
rm /run/container/check-startup /run/container/check-finish
printf 'exit\n' | container --skip-env-files --skip-all --kill-all=false
[ ! -e /run/container/check-startup ]
[ ! -e /run/container/check-finish ]
printf 'exit\n' | container --skip-env-files --exec runtime-check --skip-process --bash --kill-all=false
[ -f /run/container/check-startup ]
[ -f /run/container/check-finish ]

# The custom command must end Bash too, without waiting for EOF on stdin.
mkfifo /run/container/check-command-stdin
container --skip-env-files --skip-all --kill-all=false --bash -- sh -c 'exit 9' < /run/container/check-command-stdin &
runner=$!
exec 3> /run/container/check-command-stdin
status=0
wait "$runner" || status=$?
runner=
exec 3>&-
[ "$status" -eq 9 ]
