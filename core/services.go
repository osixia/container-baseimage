package core

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/osixia/container-baseimage/log"
)

type ServicesConfig struct {
	PriorityFilename string
	DefaultPriority  int

	InstallFilename   string
	InstalledFilename string
	StartupFilename   string
	ProcessFilename   string
	FinishFilename    string

	// for optionals services
	OptionalFilename string
	DownloadFilename string
}

func (svcsc *ServicesConfig) Validate() (bool, error) {

	if svcsc.PriorityFilename == "" {
		return false, fmt.Errorf("%v: %w", "PriorityFilename", ErrRequired)
	}

	if svcsc.InstallFilename == "" {
		return false, fmt.Errorf("%v: %w", "InstallFilename", ErrRequired)
	}
	if svcsc.InstalledFilename == "" {
		return false, fmt.Errorf("%v: %w", "InstalledFilename", ErrRequired)
	}
	if svcsc.StartupFilename == "" {
		return false, fmt.Errorf("%v: %w", "StartupFilename", ErrRequired)
	}
	if svcsc.ProcessFilename == "" {
		return false, fmt.Errorf("%v: %w", "ProcessFilename", ErrRequired)
	}
	if svcsc.FinishFilename == "" {
		return false, fmt.Errorf("%v: %w", "FinishFilename", ErrRequired)
	}

	if svcsc.OptionalFilename == "" {
		return false, fmt.Errorf("%v: %w", "OptionalFilename", ErrRequired)
	}
	if svcsc.DownloadFilename == "" {
		return false, fmt.Errorf("%v: %w", "DownloadFilename", ErrRequired)
	}

	return true, nil
}

type ServicesListOptions struct {
	Optional  *bool
	Installed *bool

	SortByPriority *bool
}

func WithOptionalServices(b bool) ServicesListOption {
	return func(slo *ServicesListOptions) {
		slo.Optional = &b
	}
}

func WithInstalledServices(b bool) ServicesListOption {
	return func(slo *ServicesListOptions) {
		slo.Installed = &b
	}
}

func SortServicesByPriority(b bool) ServicesListOption {
	return func(slo *ServicesListOptions) {
		slo.SortByPriority = &b
	}
}

type ServicesListOption func(*ServicesListOptions)

type service struct {
	name string

	dir string

	defaultPriority int
	priorityFile    string

	optionalFile string
	downloadFile string

	installFile   string
	installedFile string
	startupFile   string
	processFile   string
	finishFile    string
}

type Service interface {
	Name() string

	PriorityFile() string
	OptionalFile() string

	DownloadFile() string
	InstallFile() string
	InstalledFile() string
	StartupFile() string
	ProcessFile() string
	FinishFile() string

	InstalledFileExpectedPath() string

	Priority() int
	IsOptional() bool
	IsInstalled() bool

	IsLinkable() bool
}

type Services interface {
	Get(name string) (Service, error)
	Exists(name string) (bool, error)

	List(opts ...ServicesListOption) ([]Service, error)
	SortByPriority(services []Service)
	Join(services []Service, sep string) string

	Require(ctx context.Context, services []Service) error
	Install(ctx context.Context, services []Service) error

	Config() *ServicesConfig
}

type services struct {
	fs Filesystem

	config *ServicesConfig
}

func NewServices(fs Filesystem, svcsc *ServicesConfig) (Services, error) {
	if _, err := svcsc.Validate(); err != nil {
		return nil, err
	}

	return &services{
		fs:     fs,
		config: svcsc,
	}, nil
}

func (svcs *services) Require(ctx context.Context, services []Service) error {

	log.Tracef("services.Require called with services: %v", services)

	subCtx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	exec := NewExec(subCtx)

	for _, service := range services {

		log.Infof("Require service %v", service.Name())

		// run service download.sh script
		if service.DownloadFile() != "" {
			if err := exec.Script(service.DownloadFile()); err != nil {
				return err
			}
		}

		// remove service optional file
		if service.OptionalFile() != "" {
			log.Infof("Remove %v", service.OptionalFile())
			if err := os.Remove(service.OptionalFile()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (svcs *services) Install(ctx context.Context, services []Service) error {

	log.Tracef("services.Install called with services: %v", services)

	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	// order services by priority
	svcs.SortByPriority(services)

	// iterate services and call install script
	for _, service := range services {

		log.Debugf("Installing %v service...", service.Name())

		if service.IsInstalled() {
			log.Tracef("service %v already installed", service.Name())
			continue
		}

		// Run script if exists
		if service.InstallFile() != "" {
			if err := NewExec(ctx).Script(service.InstallFile()); err != nil {
				return err
			}
		}

		// Create .installed
		log.Debugf("Create %v", service.InstalledFileExpectedPath())
		if _, err := os.Create(service.InstalledFileExpectedPath()); err != nil {
			return err
		}
	}

	return nil
}

func (svcs *services) SortByPriority(services []Service) {

	log.Tracef("services.SortByPriority called with services: %v", services)

	priorities := map[string]int{}

	sort.Slice(services, func(i, j int) bool {
		pi, ok := priorities[services[i].Name()]
		if !ok {
			pi := services[i].Priority()
			priorities[services[i].Name()] = pi
		}

		pj, ok := priorities[services[j].Name()]
		if !ok {
			pj := services[j].Priority()
			priorities[services[j].Name()] = pj
		}

		return pi < pj
	})
}

func (svcs *services) Exists(name string) (bool, error) {

	log.Tracef("services.Exists called with name: %v", name)

	d := filepath.Join(svcs.fs.Paths().Services, name)

	if ok, err := IsDir(d); err != nil && !os.IsNotExist(err) {
		return false, err
	} else if os.IsNotExist(err) || !ok {
		return false, fmt.Errorf("%v: %w", name, ErrServiceNotFound)
	}

	return true, nil
}

func (svcs *services) List(opts ...ServicesListOption) ([]Service, error) {

	log.Tracef("services.List called with opts: %v", opts)

	slo := &ServicesListOptions{}
	for _, opt := range opts {
		opt(slo)
	}

	services := []Service{}
	servicesDir := svcs.fs.Paths().Services

	subdirs, err := os.ReadDir(servicesDir)
	if err != nil {
		return nil, err
	}
	for _, subDir := range subdirs {

		name := subDir.Name()

		if !subDir.IsDir() {
			log.Warningf("Ignoring %v in %v : not a directory", name, servicesDir)
			continue
		}

		service, err := svcs.Get(name)
		if err != nil {
			return nil, err
		}

		// filter optional services
		if slo.Optional != nil && *slo.Optional != service.IsOptional() {
			continue
		}

		// filter installed services
		if slo.Installed != nil && *slo.Installed != service.IsInstalled() {
			continue
		}

		services = append(services, service)
	}

	if slo.SortByPriority != nil && *slo.SortByPriority {
		svcs.SortByPriority(services)
	}

	return services, nil
}

func (svcs *services) Join(services []Service, sep string) string {
	names := make([]string, 0, len(services))

	for _, s := range services {
		names = append(names, s.Name())
	}

	return strings.Join(names, sep)
}

func (svcs *services) Config() *ServicesConfig {
	return svcs.config
}

func (svcs *services) Get(name string) (Service, error) {

	if _, err := svcs.Exists(name); err != nil {
		return nil, err
	}

	d := filepath.Join(svcs.fs.Paths().Services, name)

	s := &service{
		name: name,

		dir: d,

		defaultPriority: svcs.config.DefaultPriority,
		priorityFile:    filepath.Join(d, svcs.config.PriorityFilename),

		optionalFile: filepath.Join(d, svcs.config.OptionalFilename),
		downloadFile: filepath.Join(d, svcs.config.DownloadFilename),

		installFile:   filepath.Join(d, svcs.config.InstallFilename),
		installedFile: filepath.Join(d, svcs.config.InstalledFilename),
		startupFile:   filepath.Join(d, svcs.config.StartupFilename),
		processFile:   filepath.Join(d, svcs.config.ProcessFilename),
		finishFile:    filepath.Join(d, svcs.config.FinishFilename),
	}

	return s, nil
}

func (s *service) Priority() int {

	log.Trace("service.Priority called")

	file, err := os.Open(s.priorityFile)
	if err != nil {
		log.Debugf("%v file not found. Using default priority %v for service %v", s.priorityFile, s.defaultPriority, s.name)
		return s.defaultPriority
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	line, _, err := reader.ReadLine()
	if err != nil {
		log.Debugf(err.Error())
		return s.defaultPriority
	}

	priority, err := strconv.Atoi(strings.TrimSpace(string(line)))
	if err != nil {
		log.Debugf(err.Error())
		return s.defaultPriority
	}

	return priority
}

func (s *service) IsInstalled() bool {
	return s.InstalledFile() != ""
}

func (s *service) IsOptional() bool {
	return s.OptionalFile() != ""
}

func (s *service) InstalledFile() string {
	return s.file(s.installedFile)
}

func (s *service) InstalledFileExpectedPath() string {
	return s.installedFile
}

func (s *service) OptionalFile() string {
	return s.file(s.optionalFile)
}

func (s *service) PriorityFile() string {
	return s.file(s.priorityFile)
}

func (s *service) DownloadFile() string {
	return s.file(s.downloadFile)
}

func (s *service) InstallFile() string {
	return s.file(s.installFile)
}

func (s *service) StartupFile() string {
	return s.file(s.startupFile)
}

func (s *service) ProcessFile() string {
	return s.file(s.processFile)
}

func (s *service) FinishFile() string {
	return s.file(s.finishFile)
}

func (s *service) IsLinkable() bool {
	return s.ProcessFile() != "" || s.StartupFile() != "" || s.FinishFile() != ""
}

func (s *service) Name() string {
	return s.name
}

func (s *service) file(file string) string {
	if ok, _ := IsFile(file); !ok {
		return ""
	}

	return file
}
