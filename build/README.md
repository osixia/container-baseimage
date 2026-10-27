# CI/CD tool

The reusable packages now live in [osixia/dagger-build](https://github.com/osixia/dagger-build).
This directory keeps the existing CLI entrypoint, image matrix, custom build
arguments, and baseimage-specific integration tests. Requires Go 1.26.10 and
uses Dagger SDK/engine 0.21.10.

The library is pinned to a published commit in `go.mod`. Go downloads it as a
regular dependency; no sibling checkout or additional CI setup is required.

## Commands

Run from this directory:

```bash
go run main.go --help
go mod vendor
go run main.go job test ../Dockerfile --target base --export /usr/sbin/container --export-to ../bin
go run main.go job test ../Dockerfile --target nonroot --tags-suffix=-nonroot
```

Limit a local run with `--distributions debian --arches amd64`. Pass the arch of
the host: tests run on every built image, and other arches would be emulated.
The CI builds and tests each arch on its own native runner, then publishes them
together from their OCI archives (`--oci-export-to`, `--oci-import-from`).

The shared `job.DefaultTest` builds/exports images and checks their entrypoint's
version. `main.go` installs `job.Test` to run the additional checks in `tests/`:
lifecycle hooks, exit codes, restarts, signals, Bash/custom-command exit,
PID 1 process control between UIDs, actual single/multi-process example scripts,
and Nginx/PHP-FPM HTTP responses together and across containers.

Tests use derived containers, preserving the original images for publication.
Examples use their real service/environment directories through Dagger; their
Dockerfiles are not rebuilt. Web tutorials run only on root Debian/Ubuntu
images. These checks do not establish gVisor compatibility.

Run `go test -race ./...` and `go vet ./...` separately in the supervisor root,
this directory, and the shared library's checkout.
