package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/osixia/container-baseimage/log"
)

const (
	EntrypointLinkedServiceFilename = "run" // do not changed base on runit script name requirement
)

type EntrypointLinkedServicesListOption func(*EntrypointLinkedServicesListOptions)

type EntrypointLinkedServicesListOptions struct {
	Steps          []LifecycleStep
	SortByPriority *bool
}

type EntrypointLinkedService interface {
	Service() Service

	LifecycleSteps() []LifecycleStep
	Script(step LifecycleStep) string
}

type entrypointLinkedService struct {
	service Service

	scripts map[LifecycleStep]string
}

func newEntrypointLinkedService(svc Service) *entrypointLinkedService {
	return &entrypointLinkedService{
		service: svc,
		scripts: make(map[LifecycleStep]string),
	}
}

func (ls *entrypointLinkedService) Service() Service {
	return ls.service
}

func (ls *entrypointLinkedService) LifecycleSteps() []LifecycleStep {

	steps := make([]LifecycleStep, 0, len(ls.scripts))
	for step := range ls.scripts {
		steps = append(steps, step)
	}

	return steps
}

func (ls *entrypointLinkedService) Script(step LifecycleStep) string {
	return ls.scripts[step]
}

func LinkedServicesWithStep(steps ...LifecycleStep) EntrypointLinkedServicesListOption {
	return func(lso *EntrypointLinkedServicesListOptions) {
		if lso.Steps == nil {
			lso.Steps = []LifecycleStep{}
		}

		lso.Steps = append(lso.Steps, steps...)
	}
}

func LinkedServicesSortedByPriotity(b bool) EntrypointLinkedServicesListOption {
	return func(lso *EntrypointLinkedServicesListOptions) {
		lso.SortByPriority = &b
	}
}

type Entrypoint interface {
	Run(ctx context.Context, epo EntrypointOptions) (int, error)

	LinkServices(services []Service) ([]EntrypointLinkedService, error)
	UnlinkServices(linkedServices []EntrypointLinkedService) ([]Service, error)

	GetLinkedService(name string) (EntrypointLinkedService, error)
	ExistsLinkedService(name string) (bool, error)

	ListLinkedServices(opts ...EntrypointLinkedServicesListOption) ([]EntrypointLinkedService, error)
	SortByPriorityLinkedServices(linkedServices []EntrypointLinkedService)
	JoinLinkedServices(linkedServices []EntrypointLinkedService, sep string) string
}

type EntrypointOptions struct {
	SkipEnvFiles bool

	LifecycleOptions

	Services []string

	UnsecureFastWrite bool

	InstallPackages []string

	KeepAlive bool
}

func (epo *EntrypointOptions) Validate(s Services) error {

	for _, svc := range epo.Services {
		if ok, err := s.Exists(svc); err != nil {
			return err
		} else if !ok {
			services, err := s.List()
			if err != nil {
				return err
			}

			return fmt.Errorf("%v: %w (choices: %v)", svc, ErrServiceNotFound, s.Join(services, ", "))
		}
	}

	return nil
}

type entrypoint struct {
	fs   Filesystem
	dist Distribution
	svcs Services
}

func NewEntrypoint(fs Filesystem, dist Distribution, svcs Services) (Entrypoint, error) {

	return &entrypoint{
		fs:   fs,
		dist: dist,
		svcs: svcs,
	}, nil
}

func (ep *entrypoint) Run(ctx context.Context, epo EntrypointOptions) (int, error) {

	log.Trace("entrypoint.Run called")

	if err := epo.Validate(ep.svcs); err != nil {
		return 1, err
	}

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
	if err := ep.dist.InstallPackages(ctx, epo.InstallPackages); err != nil {
		return 1, err
	}

	// get services to run
	services := make([]Service, 0, len(epo.Services))
	for _, name := range epo.Services {
		s, err := ep.svcs.Get(name)
		if err != nil {
			return 1, err
		}

		services = append(services, s)
	}

	// if no service defined run all services installed
	if len(epo.Services) == 0 {
		svcs, err := ep.svcs.List(WithInstalledServices(true))
		if err != nil {
			return 1, err
		}
		services = svcs
	}

	// install services
	if err := ep.svcs.Install(ctx, services); err != nil {
		return 1, err
	}

	// link services to entrypoint
	lss, err := ep.LinkServices(services)
	if err != nil {
		return 1, err
	}

	// TODO unlink other services mais pas la multiprocess stack :/

	// check multiprocess stack is required
	isMultiprocess, err := ep.isMultiprocess()
	if err != nil {
		return 1, err
	}

	if isMultiprocess {
		if ok, err := ep.isMultiprocessStackLinked(); err != nil {
			return 1, err
		} else if !ok {
			log.Warningf("For better performances, it's recommended to install and link multiprocess stack services to entrypoint during image build.")
			if err := ep.addMultiprocessStack(ctx); err != nil {
				return 1, err
			}
		}
	}

	// set environment variables from environment files
	if epo.SkipEnvFiles {
		log.Info("Skipping getting environment variables values from environment file(s) ...")
	} else if err := ep.fs.LoadDotEnv(); err != nil {
		return 1, err
	}

	// log environment variables values
	log.Debugf("Environment variables:\n%v", strings.Join(os.Environ(), "\n"))

	// create lifecycle config based on linked services
	lcc := newLifecycleConfig(lss, ep.fs.Paths().EntrypointProcess)

	// run entrypoint lifecycle
	lc := newLifecycle(lcc, &epo.LifecycleOptions)
	exitCode, err := lc.run(ctx)
	if err != nil {
		return 1, err
	}

	// keep alive
	if epo.KeepAlive {
		log.Info("All processes have exited, keep container alive ☠ ...")
		for {
			time.Sleep(24 * time.Hour)
		}
	}

	return exitCode, nil
}

func (ep *entrypoint) LinkServices(services []Service) ([]EntrypointLinkedService, error) {

	log.Tracef("entrypoint.LinkServices called services: %v", ep.svcs.Join(services, ", "))

	linkedServices := make([]EntrypointLinkedService, 0, len(services))

	// Link startup.sh scripts of /container/services/* to /container/entrypoint/startup/*/run
	// Link process.sh scripts of /container/services/* to /container/entrypoint/process/*/run
	// Link finish.sh scripts of /container/services/* to /container/entrypoint/finish/*/run
	for _, service := range services {

		if !service.IsInstalled() {
			return nil, fmt.Errorf("service %v not installed", service.Name())
		}

		if !service.IsLinkable() {
			continue
		}

		log.Infof("Linking %v service to entrypoint ...", service.Name())

		linkedService := newEntrypointLinkedService(service)

		scripts := map[LifecycleStep]string{
			LifecycleStepStartup: service.StartupFile(),
			LifecycleStepProcess: service.ProcessFile(),
			LifecycleStepFinish:  service.FinishFile(),
		}

		for step, script := range scripts {
			if script == "" {
				continue
			}

			entrypointScript, err := ep.entrypointScript(step, service.Name())
			if err != nil {
				return nil, err
			}

			if err := Symlink(script, entrypointScript); err != nil {
				return nil, err
			}

			linkedService.scripts[step] = entrypointScript
		}

		linkedServices = append(linkedServices, linkedService)
	}

	return linkedServices, nil
}

func (ep *entrypoint) UnlinkServices(linkedServices []EntrypointLinkedService) ([]Service, error) {

	log.Tracef("entrypoint.UnlinkServices called linkedServices: %v", linkedServices)

	services := make([]Service, 0, len(linkedServices))

	for _, linkedService := range linkedServices {
		for _, step := range linkedService.LifecycleSteps() {

			script := linkedService.Script(step)

			log.Infof("Unlinking %v ...", script)
			if err := os.RemoveAll(script); err != nil {
				return nil, err
			}
		}

		services = append(services, linkedService.Service())
	}

	return services, nil
}

func (ep *entrypoint) GetLinkedService(name string) (EntrypointLinkedService, error) {

	log.Tracef("entrypoint.GetLinkedService called name: %v", name)

	service, err := ep.svcs.Get(name)
	if err != nil {
		return nil, err
	}

	linkedService := newEntrypointLinkedService(service)
	steps := []LifecycleStep{
		LifecycleStepStartup,
		LifecycleStepProcess,
		LifecycleStepFinish,
	}

	for _, step := range steps {
		entrypointScript, err := ep.entrypointScript(step, name)
		if err != nil {
			return nil, err
		}

		if _, err := os.Stat(entrypointScript); err != nil && !os.IsNotExist(err) {
			return nil, err
		} else if os.IsNotExist(err) {
			continue
		}

		linkedService.scripts[step] = entrypointScript
	}
	if len(linkedService.scripts) == 0 {
		return nil, fmt.Errorf("%v: %w", name, ErrLinkedServiceNotFound)
	}

	return linkedService, nil
}

func (ep *entrypoint) JoinLinkedServices(linkedServices []EntrypointLinkedService, sep string) string {
	names := make([]string, len(linkedServices))

	for i, s := range linkedServices {
		names[i] = s.Service().Name()
	}

	return strings.Join(names, sep)
}

func (ep *entrypoint) ExistsLinkedService(name string) (bool, error) {
	log.Tracef("entrypoint.ExistsLinkedService called with name: %v", name)

	if _, err := ep.GetLinkedService(name); err != nil {
		return false, err
	}

	return true, nil
}

func (ep *entrypoint) entrypointScript(step LifecycleStep, service string) (string, error) {

	log.Tracef("entrypoint.entrypointScript called with step: %v, service: %v", step, service)

	d, err := ep.lifecycleStepDir(step)
	if err != nil {
		return "", err
	}

	return filepath.Join(d, service, EntrypointLinkedServiceFilename), nil
}

func (ep *entrypoint) ListLinkedServices(opts ...EntrypointLinkedServicesListOption) ([]EntrypointLinkedService, error) {

	log.Tracef("entrypoint.ListLinkedServices called with opts: %v", opts)

	lso := &EntrypointLinkedServicesListOptions{}
	for _, opt := range opts {
		opt(lso)
	}

	if len(lso.Steps) == 0 {
		lso.Steps = LifecycleSteps
	}

	linkedServicesMap := make(map[string]*entrypointLinkedService)

	for _, step := range lso.Steps {

		d, err := ep.lifecycleStepDir(step)
		if err != nil {
			return nil, err
		}

		log.Tracef("Getting linked services in %v ...", d)

		err = filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				log.Tracef("Ignoring directory %v ...", path)
				return nil
			}

			if filepath.Base(path) != EntrypointLinkedServiceFilename {
				log.Tracef("Ignoring file %v ...", path)
				return nil
			}
			servicePath := filepath.Dir(path)
			name := filepath.Base(servicePath)

			if _, ok := linkedServicesMap[name]; !ok {
				service, err := ep.svcs.Get(name)
				if err != nil {
					return err
				}

				linkedServicesMap[name] = newEntrypointLinkedService(service)
			}

			linkedServicesMap[name].scripts[step] = path

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	var linkedServices []EntrypointLinkedService
	for _, linkedService := range linkedServicesMap {
		linkedServices = append(linkedServices, linkedService)
	}

	if lso.SortByPriority != nil && *lso.SortByPriority {
		ep.SortByPriorityLinkedServices(linkedServices)
	}

	log.Tracef("Linked services found: %v ...", ep.JoinLinkedServices(linkedServices, ", "))

	return linkedServices, nil
}

func (ep *entrypoint) SortByPriorityLinkedServices(linkedServices []EntrypointLinkedService) {

	log.Tracef("entrypoint.SortByPriorityLinkedServices called with linkedServices: %v", linkedServices)

	priorities := map[string]int{}

	sort.Slice(linkedServices, func(i, j int) bool {
		pi, ok := priorities[linkedServices[i].Service().Name()]
		if !ok {
			pi := linkedServices[i].Service().Priority()
			priorities[linkedServices[i].Service().Name()] = pi
		}

		pj, ok := priorities[linkedServices[j].Service().Name()]
		if !ok {
			pj := linkedServices[j].Service().Priority()
			priorities[linkedServices[j].Service().Name()] = pj
		}

		return pi < pj
	})
}

func (ep *entrypoint) lifecycleStepDir(epls LifecycleStep) (string, error) {

	log.Tracef("lifecycle.lifecycleStepDir called with epls: %v", epls)

	switch epls {
	case LifecycleStepStartup:
		return ep.fs.Paths().EntrypointStartup, nil
	case LifecycleStepProcess:
		return ep.fs.Paths().EntrypointProcess, nil
	case LifecycleStepFinish:
		return ep.fs.Paths().EntrypointFinish, nil
	}

	return "", fmt.Errorf("lifecycleStepDir: no path for %v lifecycle", epls)
}

func (ep *entrypoint) isMultiprocessStackLinked() (bool, error) {

	log.Trace("entrypoint.isMultiprocessStackLinked called")

	multiprocessServices := ep.dist.Config().MultiprocessStackServices

	missingServices := []string{}
	for _, service := range multiprocessServices {

		// check service installed
		svc, err := ep.svcs.Get(service)
		if err != nil && !IsErrServiceNotFound(err) {
			return false, err
		}

		if IsErrServiceNotFound(err) || !svc.IsInstalled() {
			missingServices = append(missingServices, service)
			continue
		}

		if !svc.IsLinkable() {
			continue
		}

		exists, err := ep.ExistsLinkedService(service)

		if err != nil && !IsErrLinkedServiceNotFound(err) {
			return false, err
		}

		if !exists {
			missingServices = append(missingServices, service)
		}
	}

	if len(missingServices) > 0 {
		if len(missingServices) == len(multiprocessServices) {
			return false, nil
		}
		return false, fmt.Errorf("%v: %w", strings.Join(missingServices, ", "), ErrMultiprocessStackPartiallyAdded)
	}

	return true, nil
}

func (ep *entrypoint) isMultiprocess() (bool, error) {

	log.Trace("entrypoint.isMultiprocess called")

	// Count image processes.
	processes, err := ep.ListLinkedServices(LinkedServicesWithStep(LifecycleStepProcess))
	if err != nil {
		return false, err
	}

	return len(processes) > 1, nil
}

func (ep *entrypoint) addMultiprocessStack(ctx context.Context) error {

	log.Trace("entrypoint.addMultiprocessStack called")

	scripts := []string{
		ep.dist.Config().BinPackagesIndexUpdate,
		ep.dist.Config().BinAddMultiprocessStack,
		ep.dist.Config().BinPackagesIndexClean,
	}

	if err := NewExec(ctx).Scripts(scripts); err != nil {
		return err
	}

	serviceNames := ep.dist.Config().MultiprocessStackServices

	services := make([]Service, 0, len(serviceNames))
	for _, serviceName := range serviceNames {
		service, err := ep.svcs.Get(serviceName)
		if err != nil {
			return err
		}
		services = append(services, service)
	}

	// Install multiprocess stack services.
	if err := ep.svcs.Install(ctx, services); err != nil {
		return err
	}

	// Link multiprocess stack services to entrypoint.
	if _, err := ep.LinkServices(services); err != nil {
		return err
	}

	return nil
}
