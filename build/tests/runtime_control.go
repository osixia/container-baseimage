package tests

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"dagger.io/dagger"
)

// RuntimeControl checks a real PID 1 supervisor from different UIDs and PID namespaces.
func RuntimeControl(ctx context.Context, client *dagger.Client, image *dagger.Container) error {

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	state := client.CacheVolume("baseimage-runtime-control-" + rand.Text())
	fixture := image.
		WithMountedCache("/run/container", state, dagger.ContainerWithMountedCacheOpts{Source: image.Directory("/run/container")}).
		WithNewFile("/container/services/runtime-control/startup.sh", "#!/bin/sh\nset -eu\n[ \"$PPID\" -eq 1 ]\n", dagger.ContainerWithNewFileOpts{Permissions: 0755}).
		WithNewFile("/container/services/runtime-control/process.sh", runtimeControlProcess, dagger.ContainerWithNewFileOpts{Permissions: 0755}).
		WithNewFile("/container/services/runtime-control/.tags/control", "").
		WithNewFile("/container/services/runtime-control/finish.sh", "#!/bin/sh\nset -eu\ntouch /run/container/control-finish\n", dagger.ContainerWithNewFileOpts{Permissions: 0755})

	supervisor := fixture.AsService(dagger.ContainerAsServiceOpts{
		Args:   []string{"container", "--skip-env-files", "--exec", "runtime-control", "--restart=false", "--kill-all-timeout", "2s"},
		NoInit: true,
	})
	if _, err := supervisor.Start(ctx); err != nil {
		return fmt.Errorf("start PID 1 supervisor: %w", err)
	}
	defer stopTestService(ctx, supervisor)

	checks := []struct {
		user   string
		script string
	}{
		{"10001:10001", `
attempt=0
while [ ! -f /run/container/control-ready ]; do
  attempt=$((attempt + 1)); [ "$attempt" -lt 100 ]; sleep 0.1
done
container processes check runtime-control
container processes stop --timeout 10 tag:control
[ -f /run/container/control-stopped ]
[ ! -e /run/container/control-finish ]
[ ! -e /run/container/processes/runtime-control.pid ]
`},
		{"10002:10002", `
container processes start --timeout 10 tag:control
attempt=0
while [ "$(wc -l < /run/container/control-attempts)" -lt 2 ]; do
  attempt=$((attempt + 1)); [ "$attempt" -lt 100 ]; sleep 0.1
done
[ "$(wc -l < /run/container/control-attempts)" -eq 2 ]
touch /run/container/control-reload-request
attempt=0
while [ ! -f /run/container/control-reloaded ]; do
  attempt=$((attempt + 1)); [ "$attempt" -lt 100 ]; sleep 0.1
done
container processes check tag:control
`},
		{"10003:10003", `
container processes status runtime-control | grep -F 'Up:true'
container processes check runtime-control
container services status runtime-control
`},
	}

	for _, check := range checks {
		if _, err := fixture.WithUser(check.user).WithExec([]string{"sh", "-eu", "-c", check.script}).Sync(ctx); err != nil {
			return fmt.Errorf("runtime control as %s: %w", check.user, err)
		}
	}

	if _, err := supervisor.Stop(ctx); err != nil {
		return fmt.Errorf("stop PID 1 supervisor: %w", err)
	}

	if _, err := fixture.WithUser("10003:10003").WithExec([]string{"sh", "-eu", "-c", `
[ -f /run/container/control-finish ]
[ ! -e /run/container/processes/runtime-control.pid ]
`}).Sync(ctx); err != nil {
		return fmt.Errorf("PID 1 finish hooks and state cleanup: %w", err)
	}

	return nil
}

func stopTestService(ctx context.Context, service *dagger.Service) {

	cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	_, _ = service.Stop(cleanup, dagger.ServiceStopOpts{Kill: true})
}

const runtimeControlProcess = `#!/bin/sh
set -eu
echo attempt >> /run/container/control-attempts
trap 'touch /run/container/control-stopped; exit 0' TERM
trap 'touch /run/container/control-reloaded' HUP
touch /run/container/control-ready
while :; do
  if [ -f /run/container/control-reload-request ]; then
    rm /run/container/control-reload-request
    kill -HUP "$PPID"
  fi
  sleep 0.05
done
`
