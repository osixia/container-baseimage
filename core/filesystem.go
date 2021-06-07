package core

import (
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/subosito/gotenv"

	"github.com/osixia/container-baseimage/log"
)

type FilesystemConfig struct {
	RootPath    string
	RunRootPath string

	EnvironmentFilesPrefix string
}

func (fsc *FilesystemConfig) Validate() (bool, error) {

	if fsc.RootPath == "" {
		return false, fmt.Errorf("%v: %w", "RootPath", ErrRequired)
	}
	if fsc.RunRootPath == "" {
		return false, fmt.Errorf("%v: %w", "RunRootPath", ErrRequired)
	}

	if fsc.EnvironmentFilesPrefix == "" {
		return false, fmt.Errorf("%v: %w", "EnvironmentFilesPrefix", ErrRequired)
	}

	return true, nil
}

type filesystemPaths struct {
	Root string

	Entrypoint        string
	EntrypointStartup string
	EntrypointProcess string
	EntrypointFinish  string

	EnvironmentFiles string

	Generator          string
	GeneratorTemplates string
	GeneratorOutput    string

	Services string

	Bin string

	RunRoot            string
	RunEntrypoint      string
	RunGeneratorOutput string
}

func newFilesystemPaths(fsc *FilesystemConfig) *filesystemPaths {

	return &filesystemPaths{
		Root: fsc.RootPath,

		Entrypoint:        fsc.RootPath + "/entrypoint", // symlink to RunEntrypoint
		EntrypointStartup: fsc.RootPath + "/entrypoint/startup",
		EntrypointProcess: fsc.RootPath + "/entrypoint/process",
		EntrypointFinish:  fsc.RootPath + "/entrypoint/finish",

		EnvironmentFiles: fsc.RootPath + "/environment",

		Generator:          fsc.RootPath + "/generator",
		GeneratorTemplates: fsc.RootPath + "/generator/templates",
		GeneratorOutput:    fsc.RootPath + "/generator/output", // symlink to RunGeneratorOutput

		Services: fsc.RootPath + "/services",

		Bin: fsc.RootPath + "/bin",

		RunRoot:            fsc.RunRootPath,
		RunEntrypoint:      fsc.RunRootPath + "/entrypoint",
		RunGeneratorOutput: fsc.RunRootPath + "/generator/output",
	}
}

type Filesystem interface {
	Create() error

	ListDotEnv() ([]string, error)
	LoadDotEnv() error

	Config() *FilesystemConfig
	Paths() *filesystemPaths
}

type filesystem struct {
	config *FilesystemConfig
	paths  *filesystemPaths
}

func NewFilesystem(fsc *FilesystemConfig) (Filesystem, error) {

	if _, err := fsc.Validate(); err != nil {
		return nil, err
	}

	fsp := newFilesystemPaths(fsc)

	return &filesystem{
		config: fsc,
		paths:  fsp,
	}, nil
}

func (fs *filesystem) Create() error {

	log.Trace("filesystem.Create called")

	dirs := []string{
		fs.paths.Root,
		fs.paths.EnvironmentFiles,
		fs.paths.Generator,
		fs.paths.GeneratorTemplates,
		fs.paths.Services,
		fs.paths.Bin,
		fs.paths.RunRoot,
		fs.paths.RunEntrypoint,
		fs.paths.RunGeneratorOutput,
	}

	for _, dir := range dirs {
		log.Tracef("Creating directory %v ...", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	worldWritableDirs := []string{fs.paths.RunRoot, fs.paths.RunGeneratorOutput}
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

	symlinks := map[string]string{
		fs.paths.RunEntrypoint:      fs.paths.Entrypoint,
		fs.paths.RunGeneratorOutput: fs.paths.GeneratorOutput,
	}

	for target, dir := range symlinks {
		if err := Symlink(target, dir); err != nil {
			return err
		}
	}

	symlinksSubdirs := []string{
		fs.paths.EntrypointStartup,
		fs.paths.EntrypointProcess,
		fs.paths.EntrypointFinish,
	}

	for _, dir := range symlinksSubdirs {
		log.Tracef("Creating directory %v ...", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

func (fs *filesystem) Config() *FilesystemConfig {
	return fs.config
}

func (fs *filesystem) Paths() *filesystemPaths {
	return fs.paths
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

	log.Debugf("Searching for %v files in %v ...", fs.config.EnvironmentFilesPrefix, fs.paths.EnvironmentFiles)
	err := filepath.Walk(fs.paths.EnvironmentFiles, func(path string, info os.FileInfo, err error) error {
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
