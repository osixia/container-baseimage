package core

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/osixia/container-baseimage/log"
)

// Entrypoint options
// =============================

type EntrypointOptions struct {
	SkipEnvFiles bool

	LifecycleOptions

	Services []Service

	FastWrite bool

	KeepAlive bool
}

// Entrypoint
// =============================

type Entrypoint interface {
	Run(ctx context.Context, epo EntrypointOptions) (int, error)
}

type entrypoint struct {
	fs   Filesystem
	svcs Services
	prcs Processes
}

func newEntrypoint(fs Filesystem, svcs Services, prcs Processes) (Entrypoint, error) {

	return &entrypoint{
		fs:   fs,
		svcs: svcs,
		prcs: prcs,
	}, nil
}

func (ep *entrypoint) Run(ctx context.Context, epo EntrypointOptions) (int, error) {

	log.Tracef("entrypoint.Run called with epo: %+v", epo)

	// set unsecure fast write
	if epo.FastWrite {
		log.Info("Unsecure fast write is enabled: setting LD_PRELOAD=libeatmydata.so")
		if err := os.Setenv("LD_PRELOAD", "libeatmydata.so"); err != nil {
			log.Error("Failed to set LD_PRELOAD environment variable")
		}
	}

	// create filesystem
	if err := ep.fs.Create(); err != nil {
		return 1, err
	}

	// cleanup stale pid files left over from a previous abrupt container stop
	ep.prcs.CleanupPIDFiles()

	// order services to run by priority
	ep.svcs.SortByPriority(epo.Services)

	// prepare services to run
	if err := ep.prepareServices(ctx, epo.Services); err != nil {
		return 1, err
	}

	// set environment variables from environment files
	if !epo.SkipEnvFiles {
		files, err := ep.fs.ListDotEnv()
		if err != nil {
			return 1, err
		}

		log.Infof("Loading environment variables from %v ...", strings.Join(files, ", "))

		if err := ep.fs.LoadDotEnv(files); err != nil {
			return 1, err
		}
	} else {
		log.Debug("Skipping getting environment variables values from environment files ...")
	}

	// log environment variables values
	log.Debugf("Environment variables:\n%v", strings.Join(os.Environ(), "\n"))

	// run entrypoint lifecycle
	lc := newLifecycle(ep.prcs, &epo.LifecycleOptions, epo.Services)

	lc.run(ctx)

	// keep alive
	if epo.KeepAlive {
		log.Info("All processes have exited, keep container alive ☠ ...")
		for {
			time.Sleep(24 * time.Hour)
		}
	}

	return lc.ExitCode(), nil
}

func (ep *entrypoint) prepareServices(ctx context.Context, services []Service) error {

	log.Tracef("entrypoint.prepareServices called with services: %v", services)

	// list services to download and install
	var servicesToDownload, servicesToInstall []Service
	for _, s := range services {
		if s.DownloadFile() != "" && !s.IsDownloaded() {
			servicesToDownload = append(servicesToDownload, s)
		}

		if s.InstallFile() != "" && !s.IsInstalled() {
			servicesToInstall = append(servicesToInstall, s)
		}
	}

	// download services
	if len(servicesToDownload) > 0 {
		if err := ep.svcs.Download(ctx, servicesToDownload); err != nil {
			return err
		}
	}

	// install services
	return ep.svcs.Install(ctx, servicesToInstall)
}
