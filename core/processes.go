package core

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

// Processes constantes
// =============================

// pid files hold a decimal pid: anything bigger is not a pid file
const pidFileMaxSize = 32

// Processes config
// =============================

type ProcessesConfig struct {
	PIDFileSuffix        string
	WantedDownFileSuffix string
	StartingFileSuffix   string
	TagsDir              string
}

func (prcsc *ProcessesConfig) Validate() (bool, error) {

	if prcsc.PIDFileSuffix == "" {
		return false, fmt.Errorf("PIDFileSuffix: %w", errors.ErrRequired)
	}

	if prcsc.WantedDownFileSuffix == "" {
		return false, fmt.Errorf("WantedDownFileSuffix: %w", errors.ErrRequired)
	}

	if prcsc.StartingFileSuffix == "" {
		return false, fmt.Errorf("StartingFileSuffix: %w", errors.ErrRequired)
	}

	if prcsc.TagsDir == "" {
		return false, fmt.Errorf("TagsDir: %w", errors.ErrRequired)
	}

	return true, nil
}

// Processes list options
// =============================

type ProcessesListOptions struct {
	Names []string

	TagPrefixInNames string

	Tags []string

	WantedDown *bool
}

func WithProcessesNames(names []string) ProcessesListOption {
	return func(prcslo *ProcessesListOptions) {
		if prcslo.Names == nil {
			prcslo.Names = []string{}
		}
		prcslo.Names = append(prcslo.Names, names...)
	}
}

func HandleProcessesTagPrefixInNames(tagPrefix string) ProcessesListOption {
	return func(prcslo *ProcessesListOptions) {
		prcslo.TagPrefixInNames = tagPrefix
	}
}

func WithProcessesTags(tags []string) ProcessesListOption {
	return func(prcslo *ProcessesListOptions) {
		if prcslo.Tags == nil {
			prcslo.Tags = []string{}
		}
		prcslo.Tags = append(prcslo.Tags, tags...)
	}
}

func WithProcessesWantedDown(b bool) ProcessesListOption {
	return func(prcslo *ProcessesListOptions) {
		prcslo.WantedDown = &b
	}
}

type ProcessesListOption func(*ProcessesListOptions)

// Processes
// =============================

type Processes interface {
	New(s Service) (Process, error)

	Get(name string) (Process, error)
	List(opts ...ProcessesListOption) ([]Process, error)
	SortByName(ss []Process)

	Start(ps []Process, timeout time.Duration) error
	Stop(ps []Process, timeout time.Duration) error

	CleanupRuntimeFiles(ss []Service)

	NewWatcher() (*fsnotify.Watcher, error)
}

type processes struct {
	fs Filesystem

	config *ProcessesConfig
}

func newProcesses(fs Filesystem, prcsc *ProcessesConfig) (Processes, error) {

	if _, err := prcsc.Validate(); err != nil {
		return nil, err
	}

	return &processes{
		fs: fs,

		config: prcsc,
	}, nil
}

func (prcs *processes) New(s Service) (Process, error) {

	log.Tracef("processes.New called with s: %v", s)

	p, err := prcs.new(s.Name())
	if err != nil {
		return nil, err
	}

	if err := helpers.ValidateNames(s.Tags()...); err != nil {
		return nil, err
	}

	if len(s.Tags()) > 0 {
		for _, dir := range []string{filepath.Dir(p.tagsDir), p.tagsDir} {
			if err := helpers.ConfigureRuntimeDirectory(dir); err != nil {
				return nil, err
			}
		}
	}

	// remove tag files no longer defined by the service: the runtime
	// directory survives a container restart
	for _, t := range p.Tags() {
		if slices.Contains(s.Tags(), t) {
			continue
		}
		if err := helpers.Remove(filepath.Join(p.tagsDir, t)); err != nil {
			log.Warningf("%v: error removing old tag %v: %v", s.Name(), t, err)
		}
	}
	p.tags = nil

	// create process tag files
	for _, t := range s.Tags() {
		tf := filepath.Join(p.tagsDir, t)
		if err := helpers.CreateMarker(tf); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (prcs *processes) Get(name string) (Process, error) {

	log.Tracef("processes.Get called with name: %v", name)

	return prcs.new(name)
}

func (prcs *processes) new(name string) (*process, error) {

	log.Tracef("processes.new called with name: %v", name)

	if err := helpers.ValidateName(name); err != nil {
		return nil, err
	}

	rf := filepath.Join(prcs.fs.Paths().RunProcesses, name)

	pf := rf + prcs.config.PIDFileSuffix
	df := rf + prcs.config.WantedDownFileSuffix
	sf := rf + prcs.config.StartingFileSuffix

	td := filepath.Join(prcs.fs.Paths().RunProcesses, name, prcs.config.TagsDir)

	return &process{
		name:           name,
		tagsDir:        td,
		pidFile:        pf,
		wantedDownFile: df,
		startingFile:   sf,
	}, nil
}

func (prcs *processes) List(opts ...ProcessesListOption) ([]Process, error) {

	log.Tracef("processes.List called with opts: %+v", opts)

	prcslo := &ProcessesListOptions{}
	for _, opt := range opts {
		opt(prcslo)
	}
	log.Tracef("processes list options %+v", prcslo)

	if (prcslo.Names != nil && len(prcslo.Names) == 0) || (prcslo.Tags != nil && len(prcslo.Tags) == 0) {
		log.Trace("empty names or tags")
		return nil, nil
	}

	// candidates
	cp := make(map[string]Process)

	// search tags in names
	if prcslo.TagPrefixInNames != "" {
		log.Tracef("search tags in name with prefix: \"%v\"", prcslo.TagPrefixInNames)
		var tags []string
		for _, name := range prcslo.Names {
			if tag, ok := strings.CutPrefix(name, prcslo.TagPrefixInNames); ok {
				tags = append(tags, tag)
			}
		}

		log.Tracef("tags found: %v", tags)

		// search processes with those tags
		tps, err := prcs.List(WithProcessesTags(tags))
		if err != nil {
			return nil, err
		}

		// add matching processes to candidates
		for _, tp := range tps {
			cp[tp.Name()] = tp
		}

		log.Tracef("tags services candidates: %v", cp)
	}

	// get services by names
	if prcslo.Names != nil {
		for _, name := range prcslo.Names {
			if prcslo.TagPrefixInNames != "" && strings.HasPrefix(name, prcslo.TagPrefixInNames) {
				continue
			}

			// if already in candidates skip
			if _, ok := cp[name]; ok {
				continue
			}

			// get service by name
			p, err := prcs.Get(name)
			if err != nil {
				return nil, err
			}

			// add service to candidates
			cp[p.Name()] = p
		}
	} else {
		// search all processes in processes directory
		processesDir := prcs.fs.Paths().RunProcesses

		files, err := os.ReadDir(processesDir)
		if err != nil {
			return nil, err
		}
		for _, file := range files {

			if file.IsDir() && prcslo.Tags == nil {
				log.Debugf("Ignoring directory %v", file)
				continue
			}

			fn := file.Name()

			var name string

			// tagged processes keep their directory while restarting
			if file.IsDir() {
				name = fn
			} else if n, ok := strings.CutSuffix(fn, prcs.config.PIDFileSuffix); ok {
				name = n
			} else if n, ok := strings.CutSuffix(fn, prcs.config.WantedDownFileSuffix); ok {
				name = n
			} else if n, ok := strings.CutSuffix(fn, prcs.config.StartingFileSuffix); ok {
				name = n
			} else {
				log.Debugf("Ignoring file %v: not matching %v, %v or %v files suffix", file, prcs.config.PIDFileSuffix, prcs.config.WantedDownFileSuffix, prcs.config.StartingFileSuffix)
				continue
			}

			// if already in candidates skip
			if _, ok := cp[name]; ok {
				continue
			}

			// get process by name: any user can create files here, an invalid
			// name must not make the whole listing fail
			p, err := prcs.Get(name)
			if err != nil {
				log.Warningf("Ignoring file %v: %v", fn, err)
				continue
			}

			cp[name] = p
		}
	}

	// filter candidates
	var ps []Process

	for _, c := range cp {

		// filter wanted up processes
		if prcslo.WantedDown != nil && *prcslo.WantedDown != c.IsWantedDown() {
			continue
		}

		// filter tags
		if prcslo.Tags != nil && !c.HasTag(prcslo.Tags...) {
			continue
		}

		ps = append(ps, c)
	}

	prcs.SortByName(ps)

	return ps, nil
}

func (prcs *processes) SortByName(ps []Process) {

	log.Tracef("processes.SortByName called with ps: %v", ps)

	sort.Slice(ps, func(i, j int) bool {
		// sort alphabetically
		return ps[i].Name() < ps[j].Name()
	})
}

func (prcs *processes) Start(ps []Process, timeout time.Duration) error {

	log.Tracef("processes.Start called with ps: %+v", ps)

	for _, p := range ps {

		log.Infof("Starting %v ...", p.Name())

		if err := p.SetWantedDown(false); err != nil {
			return err
		}
	}

	// one deadline for all processes: the timeout bounds the whole command
	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	for _, p := range ps {

		for !p.IsUp() {
			if !deadline.IsZero() && time.Now().After(deadline) {
				return fmt.Errorf("%v: timed out waiting for process to start", p.Name())
			}

			log.Debugf("Waiting %v to be up", p.Name())
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}

func (prcs *processes) Stop(ps []Process, timeout time.Duration) error {

	log.Tracef("processes.Stop called with ps: %+v", ps)

	var stopping []Process
	for _, p := range ps {

		// no pid file, not starting and no stop request: nothing to stop
		// (e.g. misspelled name)
		if !p.IsWantedDown() && !p.IsStarting() && !helpers.IsFileSafe(p.PIDFile()) {
			log.Warningf("%v: not running", p.Name())
			continue
		}

		log.Infof("Stopping %v ...", p.Name())

		if err := p.SetWantedDown(true); err != nil {
			return err
		}

		stopping = append(stopping, p)
	}

	// one deadline for all processes: the timeout bounds the whole command
	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	for _, p := range stopping {

		for !p.IsDown() {
			if !deadline.IsZero() && time.Now().After(deadline) {
				return fmt.Errorf("%v: timed out waiting for process to stop", p.Name())
			}

			log.Debugf("Waiting %v to be down", p.Name())
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}

func (prcs *processes) CleanupRuntimeFiles(ss []Service) {

	log.Tracef("processes.CleanupRuntimeFiles called with ss: %v", ss)

	// another container can share the processes directory and supervise
	// other services: only clean the files of the services run here
	var names []string
	for _, s := range ss {
		if s.ProcessFile() != "" {
			names = append(names, s.Name())
		}
	}

	processesDir := prcs.fs.Paths().RunProcesses

	files, err := os.ReadDir(processesDir)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warningf("Error reading processes directory %v: %v", processesDir, err)
		}
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fn := file.Name()

		// A stop request only lasts for the container run, like pid files.
		if name, ok := strings.CutSuffix(fn, prcs.config.WantedDownFileSuffix); ok && slices.Contains(names, name) {
			log.Infof("%v: removing stop request from a previous run", name)
			if err := helpers.Remove(filepath.Join(processesDir, fn)); err != nil {
				log.Warningf("Error removing stop request %v: %v", fn, err)
			}
			continue
		}

		// The supervisor that was waiting to restart this service is gone.
		if name, ok := strings.CutSuffix(fn, prcs.config.StartingFileSuffix); ok && slices.Contains(names, name) {
			log.Infof("%v: removing starting file from a previous run", name)
			if err := helpers.Remove(filepath.Join(processesDir, fn)); err != nil {
				log.Warningf("Error removing starting file %v: %v", fn, err)
			}
			continue
		}

		name, ok := strings.CutSuffix(fn, prcs.config.PIDFileSuffix)
		if !ok || !slices.Contains(names, name) {
			continue
		}

		// Run before the services start, at container start (PID 1): a pid file
		// of a service run here is stale. Never test its pid, it may have been
		// reused by another process in a new pid namespace.
		log.Infof("%v: removing pid file from a previous run", name)
		if err := helpers.Remove(filepath.Join(processesDir, fn)); err != nil {
			log.Warningf("Error removing pid file %v: %v", fn, err)
		}
	}
}

func (prcs *processes) NewWatcher() (*fsnotify.Watcher, error) {
	return helpers.NewWatcher(prcs.fs.Paths().RunProcesses)
}

// Processes
// =============================

type Process interface {
	Name() string

	Tags() []string
	HasTag(tags ...string) bool

	IsUp() bool
	IsDown() bool

	SetWantedDown(b bool) error
	IsWantedDown() bool

	SetStarting(b bool) error
	IsStarting() bool

	PIDFile() string
	WantedDownFile() string

	Status() string
}

type process struct {
	name string

	tags    []string
	tagsDir string

	pidFile        string
	wantedDownFile string
	startingFile   string
}

func (p *process) Name() string {
	return p.name
}

func (p *process) Tags() []string {

	if p.tags != nil {
		return p.tags
	}

	p.tags = []string{}

	files, err := os.ReadDir(p.tagsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warning(err.Error())
		}

		return p.tags
	}

	for _, file := range files {
		if !file.Type().IsRegular() {
			continue
		}
		p.tags = append(p.tags, file.Name())
	}

	return p.tags
}

func (p *process) HasTag(tags ...string) bool {

	for _, t := range tags {
		if slices.Contains(p.Tags(), t) {
			return true
		}
	}

	return false
}

// IsUp and IsDown only read the shared runtime files: the processes directory
// can be shared by several containers, a pid is only meaningful in the pid
// namespace of the supervisor that wrote it. The supervisor writes the pid
// file when the process starts and removes it once the process has exited.

// IsUp reports a pid file without a stop request.
func (p *process) IsUp() bool {

	data, err := helpers.ReadFileSafe(p.pidFile, pidFileMaxSize)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return false
	}

	return !p.IsWantedDown()
}

// IsDown reports a stop request without a pid file.
func (p *process) IsDown() bool {
	return p.IsWantedDown() && !helpers.IsFileSafe(p.pidFile)
}

func (p *process) SetWantedDown(b bool) error {

	if b {
		// already wanted down
		if p.IsWantedDown() {
			return nil
		}

		return helpers.CreateMarker(p.wantedDownFile)
	}

	// already not wanted down
	if !p.IsWantedDown() {
		return nil
	}

	// another start may have removed it meanwhile
	if err := helpers.Remove(p.wantedDownFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (p *process) IsWantedDown() bool {

	return helpers.IsFileSafe(p.wantedDownFile)
}

// SetStarting marks a process waiting to start, without pid file yet: it
// stays listed by commands with no argument.
func (p *process) SetStarting(b bool) error {

	if b {
		return helpers.CreateMarker(p.startingFile)
	}

	// already not starting
	if !p.IsStarting() {
		return nil
	}

	if err := helpers.Remove(p.startingFile); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (p *process) IsStarting() bool {

	return helpers.IsFileSafe(p.startingFile)
}

func (p *process) PIDFile() string {
	return p.pidFile
}

func (p *process) WantedDownFile() string {
	return p.wantedDownFile
}

func (p *process) Status() string {
	return fmt.Sprintf("%v - Up:%v, Starting:%v, WantedDown:%v", p.Name(), p.IsUp(), p.IsStarting(), p.IsWantedDown())
}
