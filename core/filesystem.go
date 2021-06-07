package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/subosito/gotenv"

	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/helpers"
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
	RunProcesses string
	RunGenerator string
}

func newFilesystemPaths(fsc *FilesystemConfig) *filesystemPaths {

	return &filesystemPaths{
		Root: fsc.RootPath,

		Environment:        fsc.RootPath + "/environment",
		Services:           fsc.RootPath + "/services",
		GeneratorTemplates: fsc.RootPath + "/generator/templates",

		RunRoot:      fsc.RunRootPath,
		RunProcesses: fsc.RunRootPath + "/processes",
		RunGenerator: fsc.RunRootPath + "/generator",
	}
}

// Filesystem
// =============================

type Filesystem interface {
	Create() error

	ListDotEnv() ([]string, error)
	LoadDotEnv(files []string) error

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
	}

	for _, dir := range dirs {
		log.Tracef("Creating directory %v ...", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	for _, dir := range []string{fs.paths.RunRoot, fs.paths.RunProcesses, fs.paths.RunGenerator} {
		if err := helpers.ConfigureRuntimeDirectory(dir); err != nil {
			return fmt.Errorf("securing runtime directory %q: %w", dir, err)
		}
	}

	return nil
}

func (fs *filesystem) LoadDotEnv(files []string) error {

	log.Tracef("filesystem.LoadDotEnv called with files: %v", files)

	if len(files) == 0 {
		return nil
	}

	var allEnvFiles strings.Builder
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		allEnvFiles.Write(content)
		allEnvFiles.WriteString("\n")
	}

	parsed, err := gotenv.StrictParse(strings.NewReader(allEnvFiles.String()))
	if err != nil {
		// Parser errors may include secret values from the input.
		return fmt.Errorf("invalid syntax in environment files: %s", strings.Join(files, ", "))
	}

	for key, val := range parsed {
		// Preserve existing values literally, including empty values.
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		if err := os.Setenv(key, val); err != nil {
			return err
		}
	}

	return nil
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
