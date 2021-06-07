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

	UnsecureFastWrite bool

	InstallPackages []string

	KeepAlive bool
}

// Entrypoint
// =============================

type Entrypoint interface {
	Run(ctx context.Context, epo EntrypointOptions) (int, error)
}

type entrypoint struct {
	fs   Filesystem
	dist Distribution
	svcs Services
	prcs Processes
}

func newEntrypoint(fs Filesystem, dist Distribution, svcs Services, prcs Processes) (Entrypoint, error) {

	return &entrypoint{
		fs:   fs,
		dist: dist,
		svcs: svcs,
		prcs: prcs,
	}, nil
}

func (ep *entrypoint) Run(ctx context.Context, epo EntrypointOptions) (int, error) {

	log.Tracef("entrypoint.Run called with epo: %+v", epo)

	// set unsecure fast write
	if epo.UnsecureFastWrite {
		log.Info("Unsecure fast write is enabled: setting LD_PRELOAD=libeatmydata.so")
		os.Setenv("LD_PRELOAD", "libeatmydata.so")
	}

	// create filesystem
	if err := ep.fs.Create(); err != nil {
		return 1, err
	}

	// install packages
	if err := ep.dist.InstallPackages(ctx, epo.InstallPackages, true, true); err != nil {
		return 1, err
	}

	// services to run
	services, err := ep.getServices(epo.Services)
	if err != nil {
		return 1, err
	}

	// prepare services to run
	if err := ep.prepareServices(ctx, services); err != nil {
		return 1, err
	}

	// set environment variables from environment files
	if epo.SkipEnvFiles {
		log.Info("Skipping getting environment variables values from environment file(s) ...")
	} else if err := ep.fs.LoadDotEnv(); err != nil {
		return 1, err
	}

	// log environment variables values
	log.Debugf("Environment variables:\n%v", strings.Join(os.Environ(), "\n"))

	// run entrypoint lifecycle
	lc := newLifecycle(ep.prcs, &epo.LifecycleOptions, services)

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

func (ep *entrypoint) getServices(services []Service) ([]Service, error) {

	log.Tracef("entrypoint.getServices called with services: %v", services)

	// if no service is specified run service(s) linked to the entrypoint
	if len(services) == 0 {

		log.Trace("no service is specified, search all services linked to the entrypoint")

		var err error
		services, err = ep.svcs.List(WithServicesLinked(true), SortServicesByPriority(true))
		if err != nil {
			return nil, err
		}
		log.Tracef("services linked to entrypoint: %v", services)
	}

	// sort services by priority
	ep.svcs.SortByPriority(services)

	return services, nil
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
		if err := ep.dist.UpdatePackages(ctx); err != nil {
			return err
		}

		if err := ep.svcs.Download(ctx, servicesToDownload); err != nil {
			return err
		}

		if err := ep.dist.CleanPackages(ctx); err != nil {
			return err
		}
	}

	// install services
	if err := ep.svcs.Install(ctx, servicesToInstall); err != nil {
		return err
	}

	return nil
}
