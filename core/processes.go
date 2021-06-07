package core

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

// Processes config
// =============================

type ProcessesConfig struct {
	PIDFileSuffix        string
	WantedDownFileSuffix string
	TagsDir              string
}

func (prcsc *ProcessesConfig) Validate() (bool, error) {
	if prcsc.PIDFileSuffix == "" {
		return false, fmt.Errorf("PIDFileSuffix: %w", errors.ErrRequired)
	}

	if prcsc.WantedDownFileSuffix == "" {
		return false, fmt.Errorf("WantedDownFileSuffix: %w", errors.ErrRequired)
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

	Up         *bool
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

func WithProcessesUp(b bool) ProcessesListOption {
	return func(prcslo *ProcessesListOptions) {
		prcslo.Up = &b
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

	Get(name string) Process
	List(opts ...ProcessesListOption) ([]Process, error)
	SortByName(ss []Process)

	Start(ps []Process) error
	Stop(ps []Process) error

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

	p := prcs.new(s.Name())

	// create process tag files
	for _, t := range s.Tags() {
		tf := filepath.Join(p.tagsDir, t)
		if _, err := helpers.Create(tf); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (prcs *processes) Get(name string) Process {

	log.Tracef("processes.Get called with name: %v", name)

	return prcs.new(name)
}

func (prcs *processes) new(name string) *process {
	rf := filepath.Join(prcs.fs.Paths().RunProcess, name)

	pf := rf + prcs.config.PIDFileSuffix
	df := rf + prcs.config.WantedDownFileSuffix

	td := filepath.Join(prcs.fs.Paths().RunProcess, name, prcs.config.TagsDir)

	return &process{
		name:           name,
		tagsDir:        td,
		pidFile:        pf,
		wantedDownFile: df,
	}
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
			if !strings.HasPrefix(name, prcslo.TagPrefixInNames) {
				continue
			}

			tag := strings.TrimLeft(name, prcslo.TagPrefixInNames)
			tags = append(tags, tag)
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

			// if already in candidates skip
			if _, ok := cp[name]; ok {
				continue
			}

			// get service by name
			s := prcs.Get(name)

			// add service to candidates
			cp[s.Name()] = s
		}
	} else {
		// search all processes in processes directory
		processesDir := prcs.fs.Paths().RunProcess

		files, err := os.ReadDir(processesDir)
		if err != nil {
			return nil, err
		}
		for _, file := range files {

			if file.IsDir() {
				log.Infof("Ignoring directory %v", file)
				continue
			}

			fn := file.Name()

			var name string

			// match pid prefix
			if strings.HasSuffix(fn, prcs.config.PIDFileSuffix) {
				name = strings.TrimRight(fn, prcs.config.PIDFileSuffix)
			} else if strings.HasSuffix(fn, prcs.config.WantedDownFileSuffix) {
				name = strings.TrimRight(fn, prcs.config.WantedDownFileSuffix)
			} else {
				log.Infof("Ignoring file %v: not matching %v or %v files suffix", file, prcs.config.PIDFileSuffix, prcs.config.WantedDownFileSuffix)
				continue
			}

			if name == "" {
				log.Warningf("Failed to get name from filename %v", fn)
				continue
			}

			// if already in candidates skip
			if _, ok := cp[name]; ok {
				continue
			}

			// get process by name
			cp[name] = prcs.Get(name)
		}
	}

	// filter candidates
	var ps []Process

	for _, c := range cp {

		// filter up processes
		if prcslo.Up != nil && *prcslo.Up != c.IsUp() {
			continue
		}

		// filter wanted up processes
		if prcslo.WantedDown != nil && *prcslo.WantedDown != c.IsWantedDown() {
			continue
		}

		// filter tags
		if prcslo.Tags != nil && c.HasTag(prcslo.Tags...) {
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

func (prcs *processes) Start(ps []Process) error {

	log.Tracef("processes.Start called with ps: %+v", ps)

	for _, p := range ps {

		log.Infof("Starting %v ...", p.Name())

		if err := p.SetWantedDown(false); err != nil {
			return err
		}

		for !p.IsUp() {
			log.Debugf("Waiting %v to be up", p.Name())
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}

func (prcs *processes) Stop(ps []Process) error {

	log.Tracef("processes.Stop called with ps: %+v", ps)

	for _, p := range ps {

		log.Infof("Stoping %v ...", p.Name())

		if err := p.SetWantedDown(true); err != nil {
			return err
		}

		for p.IsUp() {
			log.Debugf("Waiting %v to be down", p.Name())
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}

func (prcs *processes) NewWatcher() (*fsnotify.Watcher, error) {
	return helpers.NewFSWatcher(prcs.fs.Paths().RunProcess)
}

// Processes
// =============================

type Process interface {
	Name() string

	Tags() []string
	HasTag(tags ...string) bool

	IsUp() bool

	SetWantedDown(b bool) error
	IsWantedDown() bool

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
}

func (p *process) Name() string {
	return p.name
}

func (p *process) Tags() []string {
	if p.tags != nil {
		return p.tags
	}

	files, err := os.ReadDir(p.tagsDir)
	if err != nil {
		log.Fatal(err.Error())
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

func (p *process) IsUp() bool {
	ok, _ := helpers.IsFile(p.pidFile)
	return ok
}

func (p *process) SetWantedDown(b bool) error {

	if b {
		// already wanted down
		if p.IsWantedDown() {
			return nil
		}

		_, err := helpers.Create(p.wantedDownFile)
		return err
	}

	// already not wanted down
	if !p.IsWantedDown() {
		return nil
	}

	return helpers.Remove(p.wantedDownFile)
}

func (p *process) IsWantedDown() bool {
	ok, _ := helpers.IsFile(p.wantedDownFile)
	return ok
}

func (p *process) PIDFile() string {
	return p.pidFile
}

func (p *process) WantedDownFile() string {
	return p.wantedDownFile
}

func (p *process) Status() string {
	return fmt.Sprintf("%v - Up:%v, WantedDown:%v", p.Name(), p.IsUp(), p.IsWantedDown())
}
