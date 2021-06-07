package job

import (
	"context"
	"fmt"
	"time"

	"dagger.io/dagger"
)

// Check derived containers so test fixtures never become published image layers.
func testRuntime(ctx context.Context, image *dagger.Container) error {

	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	fixture := image.
		WithNewFile("/container/services/runtime-check/startup.sh", "#!/bin/sh\nset -eu\ntouch /run/container/check-startup\n", dagger.ContainerWithNewFileOpts{Permissions: 0755}).
		WithNewFile("/container/services/runtime-check/finish.sh", "#!/bin/sh\nset -eu\ntouch /run/container/check-finish\n", dagger.ContainerWithNewFileOpts{Permissions: 0755}).
		WithNewFile("/container/services/runtime-check/process.sh", runtimeProcess, dagger.ContainerWithNewFileOpts{Permissions: 0755}).
		WithNewFile("/tmp/baseimage-runtime-check.sh", runtimeCheck, dagger.ContainerWithNewFileOpts{Permissions: 0755})

	if _, err := fixture.WithExec([]string{"sh", "/tmp/baseimage-runtime-check.sh"}).Sync(ctx); err != nil {
		return fmt.Errorf("runtime integration checks: %w", err)
	}

	return nil
}

const runtimeProcess = `#!/bin/sh
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
esac
`

const runtimeCheck = `#!/bin/sh
set -eu
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
trap 'kill -TERM "$runner" 2>/dev/null || true' EXIT
attempt=0
while [ ! -f /run/container/check-ready ]; do
 attempt=$((attempt + 1))
 [ "$attempt" -lt 100 ]
 sleep 0.1
done
kill -TERM "$runner"
wait "$runner"
trap - EXIT
[ -f /run/container/check-finish ]
rm /run/container/check-finish

export BASEIMAGE_TEST_MODE=restart
container --skip-env-files --kill-all=false --restart=true --exec runtime-check
[ "$(wc -l < /run/container/check-attempts)" -eq 2 ]
[ -f /run/container/check-finish ]

status=0
container --skip-env-files --skip-all --kill-all=false -- sh -c 'exit 9' || status=$?
[ "$status" -eq 9 ]
`
