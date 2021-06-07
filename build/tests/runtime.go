package tests

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"dagger.io/dagger"
)

// Runtime checks derived containers so fixtures never become published image layers.
func Runtime(ctx context.Context, image *dagger.Container) error {

	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
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

//go:embed testdata/runtime-process.sh
var runtimeProcess string

//go:embed testdata/runtime.sh
var runtimeCheck string
