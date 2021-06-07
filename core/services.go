package core

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

// Services config
// =============================

type ServicesConfig struct {
	TagsDir string

	DefaultPriority  int
	PriorityFilename string

	OptionalFilename string

	DownloadFilename   string
	DownloadedFilename string

	InstallFilename   string
	InstalledFilename string

	StartupFilename string
	ProcessFilename string
	FinishFilename  string

	LinkedFilename string
}

func (svcsc *ServicesConfig) Validate() (bool, error) {
	if svcsc.TagsDir == "" {
		return false, fmt.Errorf("TagDir: %w", errors.ErrRequired)
	}

	if svcsc.PriorityFilename == "" {
		return false, fmt.Errorf("PriorityFilename: %w", errors.ErrRequired)
	}

	if svcsc.OptionalFilename == "" {
		return false, fmt.Errorf("OptionalFilename: %w", errors.ErrRequired)
	}

	if svcsc.DownloadFilename == "" {
		return false, fmt.Errorf("DownloadFilename: %w", errors.ErrRequired)
	}
	if svcsc.DownloadedFilename == "" {
		return false, fmt.Errorf("DownloadedFilename: %w", errors.ErrRequired)
	}

	if svcsc.InstallFilename == "" {
		return false, fmt.Errorf("InstallFilename: %w", errors.ErrRequired)
	}
	if svcsc.InstalledFilename == "" {
		return false, fmt.Errorf("InstalledFilename: %w", errors.ErrRequired)
	}

	if svcsc.StartupFilename == "" {
		return false, fmt.Errorf("StartupFilename: %w", errors.ErrRequired)
	}
	if svcsc.ProcessFilename == "" {
		return false, fmt.Errorf("ProcessFilename: %w", errors.ErrRequired)
	}
	if svcsc.FinishFilename == "" {
		return false, fmt.Errorf("FinishFilename: %w", errors.ErrRequired)
	}

	if svcsc.LinkedFilename == "" {
		return false, fmt.Errorf("LinkedFilename: %w", errors.ErrRequired)
	}

	return true, nil
}

// Services list options
// =============================

type ServicesListOptions struct {
	Names []string

	TagPrefixInNames string

	Tags []string

	Optional   *bool
	Downloaded *bool
	Installed  *bool
	Linked     *bool

	SortByPriority *bool
}

func WithServicesNames(names []string) ServicesListOption {
	return func(svcslo *ServicesListOptions) {
		if svcslo.Names == nil {
			svcslo.Names = []string{}
		}
		svcslo.Names = append(svcslo.Names, names...)
	}
}

func HandleServicesTagPrefixInNames(tagPrefix string) ServicesListOption {
	return func(svcslo *ServicesListOptions) {
		svcslo.TagPrefixInNames = tagPrefix
	}
}

func WithServicesTags(tags []string) ServicesListOption {
	return func(svcslo *ServicesListOptions) {
		if svcslo.Tags == nil {
			svcslo.Tags = []string{}
		}
		svcslo.Tags = append(svcslo.Tags, tags...)
	}
}

func WithServicesOptional(b bool) ServicesListOption {
	return func(svcslo *ServicesListOptions) {
		svcslo.Optional = &b
	}
}

func WithServicesDownloaded(b bool) ServicesListOption {
	return func(svcslo *ServicesListOptions) {
		svcslo.Downloaded = &b
	}
}

func WithServicesInstalled(b bool) ServicesListOption {
	return func(svcslo *ServicesListOptions) {
		svcslo.Installed = &b
	}
}

func WithServicesLinked(b bool) ServicesListOption {
	return func(svcslo *ServicesListOptions) {
		svcslo.Linked = &b
	}
}

func SortServicesByPriority(b bool) ServicesListOption {
	return func(svcslo *ServicesListOptions) {
		svcslo.SortByPriority = &b
	}
}

type ServicesListOption func(*ServicesListOptions)

// Services
// =============================

type Services interface {
	Get(name string) (Service, error)

	Exists(name string) (bool, error)

	List(opts ...ServicesListOption) ([]Service, error)

	SortByName(ss []Service)
	SortByPriority(ss []Service)

	Require(ss []Service) error
	Download(ctx context.Context, ss []Service) error
	Install(ctx context.Context, ss []Service) error
	Link(ss []Service) error
	Unlink(ss []Service) error

	Config() *ServicesConfig
}

type services struct {
	fs Filesystem

	config *ServicesConfig
}

func newServices(fs Filesystem, svcsc *ServicesConfig) (Services, error) {

	if _, err := svcsc.Validate(); err != nil {
		return nil, err
	}

	return &services{
		fs:     fs,
		config: svcsc,
	}, nil
}

func (svcs *services) Get(name string) (Service, error) {

	log.Tracef("services.Get called with name: %v", name)

	if _, err := svcs.Exists(name); err != nil {
		return nil, err
	}

	d := filepath.Join(svcs.fs.Paths().Services, name)

	s := &service{
		name: name,

		tagsDir: filepath.Join(d, svcs.config.TagsDir),

		defaultPriority: svcs.config.DefaultPriority,
		priorityFile:    filepath.Join(d, svcs.config.PriorityFilename),

		optionalFile: filepath.Join(d, svcs.config.OptionalFilename),

		downloadFile:   filepath.Join(d, svcs.config.DownloadFilename),
		downloadedFile: filepath.Join(d, svcs.config.DownloadedFilename),

		installFile:   filepath.Join(d, svcs.config.InstallFilename),
		installedFile: filepath.Join(d, svcs.config.InstalledFilename),

		startupFile: filepath.Join(d, svcs.config.StartupFilename),
		processFile: filepath.Join(d, svcs.config.ProcessFilename),
		finishFile:  filepath.Join(d, svcs.config.FinishFilename),

		linkedFile: filepath.Join(d, svcs.config.LinkedFilename),
	}

	return s, nil
}

func (svcs *services) Exists(name string) (bool, error) {

	log.Tracef("services.Exists called with name: %v", name)

	d := filepath.Join(svcs.fs.Paths().Services, name)

	if ok, err := helpers.IsDir(d); err != nil && !os.IsNotExist(err) {
		return false, err
	} else if os.IsNotExist(err) || !ok {
		return false, fmt.Errorf("%v: service %w", name, errors.ErrUnknown)
	}

	return true, nil
}

func (svcs *services) List(opts ...ServicesListOption) ([]Service, error) {

	log.Tracef("services.List called with opts: %+v", opts)

	svcslo := &ServicesListOptions{}
	for _, opt := range opts {
		opt(svcslo)
	}
	log.Tracef("services list options %+v", svcslo)

	if (svcslo.Names != nil && len(svcslo.Names) == 0) || (svcslo.Tags != nil && len(svcslo.Tags) == 0) {
		log.Trace("empty names or tags")
		return nil, nil
	}

	// candidates
	cs := make(map[string]Service)

	// search tags in names
	if svcslo.TagPrefixInNames != "" {
		log.Tracef("search tags in name with prefix: \"%v\"", svcslo.TagPrefixInNames)
		var tags []string
		for _, name := range svcslo.Names {
			if !strings.HasPrefix(name, svcslo.TagPrefixInNames) {
				continue
			}

			tag := strings.TrimLeft(name, svcslo.TagPrefixInNames)
			tags = append(tags, tag)
		}
		log.Tracef("tags found: %v", tags)

		// search services with those tags
		tss, err := svcs.List(WithServicesTags(tags))
		if err != nil {
			return nil, err
		}

		// add matching services to candidates
		for _, ts := range tss {
			cs[ts.Name()] = ts
		}

		log.Tracef("tags services candidates: %v", cs)
	}

	// get services by names
	if svcslo.Names != nil {
		for _, name := range svcslo.Names {

			// not realy a name but a tag that must have been already be handle (skip)
			if svcslo.TagPrefixInNames != "" && strings.HasPrefix(name, svcslo.TagPrefixInNames) {
				continue
			}

			// if already in candidates (skip)
			if _, ok := cs[name]; ok {
				continue
			}

			// get service by name
			s, err := svcs.Get(name)
			if err != nil {
				return nil, err
			}

			// add service to candidates
			cs[s.Name()] = s
		}
	} else {
		// search all services in services directory
		servicesDir := svcs.fs.Paths().Services

		subdirs, err := os.ReadDir(servicesDir)
		if err != nil {
			return nil, err
		}
		for _, subDir := range subdirs {

			if !subDir.IsDir() {
				log.Infof("Ignoring %v: not a directory", subDir)
				continue
			}

			name := subDir.Name()

			// if already in candidates (skip)
			if _, ok := cs[name]; ok {
				continue
			}

			// get service by name
			s, err := svcs.Get(name)
			if err != nil {
				return nil, err
			}

			// add service to candidates
			cs[s.Name()] = s
		}
	}

	// filter candidates
	var ss []Service

	for _, c := range cs {

		// filter linked services
		if svcslo.Linked != nil && *svcslo.Linked != c.IsLinked() {
			continue
		}

		// filter installed services
		if svcslo.Installed != nil && *svcslo.Installed != c.IsInstalled() {
			continue
		}

		// filter downloaded services
		if svcslo.Downloaded != nil && *svcslo.Downloaded != c.IsDownloaded() {
			continue
		}

		// filter optional services
		if svcslo.Optional != nil && *svcslo.Optional != c.IsOptional() {
			continue
		}

		// filter tags
		if svcslo.Tags != nil && !c.HasTag(svcslo.Tags...) {
			continue
		}

		ss = append(ss, c)
	}

	// sort services
	if svcslo.SortByPriority != nil && *svcslo.SortByPriority {
		svcs.SortByPriority(ss)
	} else {
		svcs.SortByName(ss)
	}

	return ss, nil
}

func (svcs *services) SortByPriority(ss []Service) {

	log.Tracef("services.SortByPriority called with ss: %v", ss)

	sort.Slice(ss, func(i, j int) bool {
		// if the priorities are not equals: sort by priority
		if ss[i].Priority() != ss[j].Priority() {
			return ss[i].Priority() < ss[j].Priority()
		}
		// else sort alphabetically
		return ss[i].Name() < ss[j].Name()
	})
}

func (svcs *services) SortByName(ss []Service) {

	log.Tracef("services.SortByName called with ss: %v", ss)

	sort.Slice(ss, func(i, j int) bool {
		// sort alphabetically
		return ss[i].Name() < ss[j].Name()
	})
}

func (svcs *services) Require(ss []Service) error {

	log.Tracef("services.Require called with ss: %v", ss)

	for _, s := range ss {
		if err := s.Require(); err != nil {
			return err
		}
	}

	return nil
}

func (svcs *services) Download(ctx context.Context, ss []Service) error {

	log.Tracef("services.Download called with ss: %v", ss)

	subCtx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	for _, s := range ss {
		if err := s.Download(subCtx); err != nil {
			return err
		}
	}

	return nil
}

func (svcs *services) Install(ctx context.Context, ss []Service) error {

	log.Tracef("services.Install called with ss: %v", ss)

	subCtx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	for _, s := range ss {
		if err := s.Install(subCtx); err != nil {
			return err
		}
	}

	return nil
}

func (svcs *services) Link(ss []Service) error {

	log.Tracef("services.Link called with ss: %v", ss)

	for _, s := range ss {
		if err := s.Link(); err != nil {
			return err
		}
	}

	return nil
}

func (svcs *services) Unlink(ss []Service) error {

	log.Tracef("services.Unlink called with ss: %v", ss)

	for _, s := range ss {
		if err := s.Unlink(); err != nil {
			return err
		}
	}

	return nil
}

func (svcs *services) Config() *ServicesConfig {
	return svcs.config
}

// Service
// =============================

type Service interface {
	Name() string

	Tags() []string
	HasTag(tags ...string) bool

	Priority() int

	Require() error
	IsOptional() bool

	Download(ctx context.Context) error
	DownloadFile() string
	IsDownloaded() bool

	Install(ctx context.Context) error
	InstallFile() string
	IsInstalled() bool

	StartupFile() string
	ProcessFile() string
	FinishFile() string

	Link() error
	Unlink() error
	IsLinked() bool

	Status() string
}

type service struct {
	name string

	tags    []string
	tagsDir string

	defaultPriority int

	priority     *int
	priorityFile string

	optionalFile string

	downloadFile   string
	downloadedFile string

	installFile   string
	installedFile string

	startupFile string
	processFile string
	finishFile  string

	linkedFile string
}

func (s *service) Name() string {
	return s.name
}

func (s *service) Tags() []string {
	if s.tags != nil {
		return s.tags
	}

	files, err := os.ReadDir(s.tagsDir)
	if os.IsNotExist(err) {
		s.tags = []string{}
		return s.tags
	} else if err != nil {
		log.Fatal(err.Error())
	}

	for _, file := range files {
		if !file.Type().IsRegular() {
			continue
		}
		s.tags = append(s.tags, file.Name())
	}

	return s.tags
}

func (s *service) HasTag(tags ...string) bool {
	for _, t := range tags {
		if slices.Contains(s.Tags(), t) {
			return true
		}
	}

	return false
}

func (s *service) Priority() int {

	log.Trace("service.Priority called")

	if s.priority != nil {
		return *s.priority
	}

	file, err := os.Open(s.priorityFile)
	if err != nil {
		log.Debugf("%v file not found. Using default priority %v for service %v", s.priorityFile, s.defaultPriority, s.name)
		return s.defaultPriority
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	line, _, err := reader.ReadLine()
	if err != nil {
		log.Debug(err.Error())
		return s.defaultPriority
	}

	priority, err := strconv.Atoi(strings.TrimSpace(string(line)))
	if err != nil {
		log.Debug(err.Error())
		return s.defaultPriority
	}

	s.priority = &priority

	return *s.priority
}

func (s *service) Require() error {

	log.Trace("service.Require called")

	log.Infof("Requiring %v service ...", s.Name())

	if !s.IsOptional() {
		log.Warningf("Service %v is not marked as optional", s.Name())
		return nil
	}

	if err := helpers.Remove(s.optionalFile); err != nil {
		return err
	}

	return nil
}

func (s *service) IsOptional() bool {
	ok, _ := helpers.IsFile(s.optionalFile)
	return ok
}

func (s *service) Download(ctx context.Context) error {

	log.Trace("service.Download called")

	log.Infof("Downloading %v service ...", s.Name())

	if s.IsDownloaded() {
		log.Warningf("Service %v already downloaded", s.Name())
	}

	if s.DownloadFile() == "" {
		return nil
	}

	if err := helpers.NewExec(ctx).Command(s.DownloadFile()).Run(); err != nil {
		return err
	}

	if _, err := helpers.Create(s.downloadedFile); err != nil {
		return err
	}

	return nil
}

func (s *service) DownloadFile() string {
	return s.filePath(s.downloadFile)
}

func (s *service) IsDownloaded() bool {
	ok, _ := helpers.IsFile(s.downloadedFile)
	return ok
}

func (s *service) Install(ctx context.Context) error {

	log.Trace("service.Install called")

	log.Infof("Installing %v service...", s.Name())

	if s.IsInstalled() {
		log.Warningf("Service %v already installed", s.Name())
	}

	if s.InstallFile() == "" {
		return nil
	}

	if err := helpers.NewExec(ctx).Command(s.InstallFile()).Run(); err != nil {
		return err
	}

	if _, err := helpers.Create(s.installedFile); err != nil {
		return err
	}

	return nil
}

func (s *service) InstallFile() string {
	return s.filePath(s.installFile)
}

func (s *service) IsInstalled() bool {
	ok, _ := helpers.IsFile(s.installedFile)
	return ok
}

func (s *service) StartupFile() string {
	return s.filePath(s.startupFile)
}

func (s *service) ProcessFile() string {
	return s.filePath(s.processFile)
}

func (s *service) FinishFile() string {
	return s.filePath(s.finishFile)
}

func (s *service) Link() error {

	log.Trace("service.Link called")

	log.Infof("Linking %v service to entrypoint ...", s.Name())

	if s.IsLinked() {
		log.Warningf("Service %v already linked", s.Name())
		return nil
	}

	if _, err := helpers.Create(s.linkedFile); err != nil {
		return err
	}

	return nil
}

func (s *service) Unlink() error {

	log.Trace("service.Unlink called")

	log.Infof("Unlinking %v service to entrypoint ...", s.Name())

	if !s.IsLinked() {
		log.Warningf("Service %v not linked", s.Name())
		return nil
	}

	if err := helpers.Remove(s.linkedFile); err != nil {
		return err
	}

	return nil
}

func (s *service) IsLinked() bool {
	ok, _ := helpers.IsFile(s.linkedFile)
	return ok
}

func (s *service) Status() string {
	return fmt.Sprintf("%v - Optional:%v, Downloaded:%v, Installed:%v, Linked:%v, Tags:%v", s.Name(), s.IsOptional(), s.IsDownloaded(), s.IsInstalled(), s.IsLinked(), s.Tags())
}

func (s *service) filePath(file string) string {
	if ok, _ := helpers.IsFile(file); !ok {
		return ""
	}

	return file
}
