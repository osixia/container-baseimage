package core

import (
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/subosito/gotenv"

	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/log"
)

// Filesystem config
// =============================

type FilesystemConfig struct {
	RootPath    string
	RunRootPath string

	EnvironmentFilesPrefix string
}

func (fsc *FilesystemConfig) Validate() (bool, error) {

	if fsc.RootPath == "" {
		return false, fmt.Errorf("RootPath: %w", errors.ErrRequired)
	}
	if fsc.RunRootPath == "" {
		return false, fmt.Errorf("RunRootPath: %w", errors.ErrRequired)
	}

	if fsc.EnvironmentFilesPrefix == "" {
		return false, fmt.Errorf("EnvironmentFilesPrefix: %w", errors.ErrRequired)
	}

	return true, nil
}

// Filesystem paths
// =============================

type filesystemPaths struct {
	Root string

	Environment        string
	Services           string
	GeneratorTemplates string

	RunRoot      string
	RunProcess   string
	RunGenerator string
}

func newFilesystemPaths(fsc *FilesystemConfig) *filesystemPaths {

	return &filesystemPaths{
		Root: fsc.RootPath,

		Environment:        fsc.RootPath + "/environment",
		Services:           fsc.RootPath + "/services",
		GeneratorTemplates: fsc.RootPath + "/generator/templates",

		RunRoot:      fsc.RunRootPath,
		RunProcess:   fsc.RunRootPath + "/process",
		RunGenerator: fsc.RunRootPath + "/generator",
	}
}

// Filesystem
// =============================

type Filesystem interface {
	Create() error

	ListDotEnv() ([]string, error)
	LoadDotEnv() error

	Paths() *filesystemPaths
	Config() *FilesystemConfig
}

type filesystem struct {
	paths  *filesystemPaths
	config *FilesystemConfig
}

func newFilesystem(fsc *FilesystemConfig) (Filesystem, error) {

	if _, err := fsc.Validate(); err != nil {
		return nil, err
	}

	fsp := newFilesystemPaths(fsc)

	return &filesystem{
		paths:  fsp,
		config: fsc,
	}, nil
}

func (fs *filesystem) Create() error {

	log.Trace("filesystem.Create called")

	log.Debug("Creating container filesystem ...")

	dirs := []string{
		fs.paths.Root,
		fs.paths.Environment,
		fs.paths.GeneratorTemplates,
		fs.paths.Services,
		fs.paths.RunRoot,
		fs.paths.RunProcess,
		fs.paths.RunGenerator,
	}

	for _, dir := range dirs {
		log.Tracef("Creating directory %v ...", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	worldWritableDirs := []string{fs.paths.RunRoot, fs.paths.RunProcess, fs.paths.RunGenerator}
	var worldWritablePerm iofs.FileMode = 0777
	for _, dir := range worldWritableDirs {

		fi, err := os.Stat(dir)
		if err != nil {
			return err
		}

		if fi.Mode().Perm() != worldWritablePerm {
			log.Tracef("Setting %v permissions to %v", worldWritablePerm, dir)
			if err := os.Chmod(dir, worldWritablePerm); err != nil {
				log.Warning(err.Error())
			}
		}

	}

	return nil
}

func (fs *filesystem) LoadDotEnv() error {

	log.Trace("filesystem.LoadDotEnv called")

	files, err := fs.ListDotEnv()
	if err != nil {
		return nil
	}

	if len(files) == 0 {
		return nil
	}

	log.Infof("Loading environment variables from %v ...", strings.Join(files, ", "))
	envBackup := os.Environ()
	if err := gotenv.OverLoad(files...); err != nil {
		return err
	}

	return gotenv.OverApply(strings.NewReader(strings.Join(envBackup, "\n")))
}

func (fs *filesystem) ListDotEnv() ([]string, error) {

	log.Trace("filesystem.ListDotEnv called")

	var files []string

	log.Debugf("Searching for %v files in %v ...", fs.config.EnvironmentFilesPrefix, fs.paths.Environment)
	err := filepath.Walk(fs.paths.Environment, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasPrefix(info.Name(), fs.config.EnvironmentFilesPrefix) {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (fs *filesystem) Paths() *filesystemPaths {
	return fs.paths
}

func (fs *filesystem) Config() *FilesystemConfig {
	return fs.config
}
