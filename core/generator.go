package core

import (
	goerrors "errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

// Generator config
// =============================

type GeneratorConfig struct {
	fromImage string
}

func (genc *GeneratorConfig) Validate() (bool, error) {

	if genc.fromImage == "" {
		return false, fmt.Errorf("fromImage: %w", errors.ErrRequired)
	}

	return true, nil
}

// Generator options
// =============================

type GenerateBootstrapOptions struct {
	GenerateDockerfileOptions
	GenerateServicesOptions

	Multiprocess bool
}

type GenerateDockerfileOptions struct {
	Image string
}

type GenerateServicesOptions struct {
	Names    []string
	Priority int
	Optional bool
	Tags     []string
}

// Generator
// =============================

type Generator interface {
	GenerateBootstrap(gopt *GenerateBootstrapOptions) ([]string, error)

	GenerateDockerfile(gopt *GenerateDockerfileOptions) ([]string, error)
	GenerateEnvironment() ([]string, error)
	GenerateServices(gopt *GenerateServicesOptions) ([]string, error)

	Config() *GeneratorConfig
}

type generator struct {
	env  Environment
	fs   Filesystem
	svcs Services

	envBackup []string

	config *GeneratorConfig
}

func newGenerator(env Environment, fs Filesystem, svcs Services, genc *GeneratorConfig) (Generator, error) {

	if _, err := genc.Validate(); err != nil {
		return nil, err
	}

	return &generator{
		env:    env,
		fs:     fs,
		svcs:   svcs,
		config: genc,
	}, nil
}

func (gen *generator) GenerateBootstrap(gopt *GenerateBootstrapOptions) ([]string, error) {

	log.Tracef("generator.GenerateBootstrap called with gopt: %v", gopt)

	if len(gopt.Names) > 1 {
		gopt.Multiprocess = true
	}

	if gopt.Multiprocess {
		for len(gopt.Names) < 2 {
			gopt.Names = append(gopt.Names, fmt.Sprintf("service-%v", len(gopt.Names)+1))
		}
	}

	var files []string

	// add dockerfile
	df, err := gen.GenerateDockerfile(&gopt.GenerateDockerfileOptions)
	if err != nil {
		return nil, err
	}
	files = append(files, df...)

	// add services
	sf, err := gen.GenerateServices(&gopt.GenerateServicesOptions)
	if err != nil {
		return nil, err
	}
	files = append(files, sf...)

	// add environment
	ef, err := gen.GenerateEnvironment()
	if err != nil {
		return nil, err
	}
	files = append(files, ef...)

	return files, nil
}

func (gen *generator) GenerateDockerfile(gopt *GenerateDockerfileOptions) ([]string, error) {

	log.Trace("generator.GenerateDockerfile called")

	envDir, err := gen.envDir(0)
	if err != nil {
		return nil, err
	}

	gen.backupEnv()
	defer gen.restoreEnv()

	gen.setEnv(map[string]string{
		"FROM_IMAGE":                       gen.config.fromImage,
		"IMAGE":                            gopt.Image,
		"CONTAINER_IMAGE_ENV_KEY":          gen.env.Config().ImageKey,
		"DOCKERFILE_SERVICES_DIR":          gen.dockerfileServicesDir(),
		"CONTAINER_SERVICES_DIR":           gen.fs.Paths().Services,
		"DOCKERFILE_ENVIRONMENT_FILES_DIR": gen.dockerfileEnvDir(),
		"CONTAINER_ENVIRONMENT_FILES_DIR":  envDir,
	})

	// add dockerfile
	return gen.output("Dockerfile.template", "Dockerfile.template")
}

func (gen *generator) GenerateEnvironment() ([]string, error) {

	log.Tracef("generator.GenerateEnvironment called")

	gen.backupEnv()
	defer gen.restoreEnv()

	gen.setEnv(nil)

	// add environment
	return gen.output("/environment", gen.dockerfileEnvDir())
}

func (gen *generator) GenerateServices(gopt *GenerateServicesOptions) ([]string, error) {

	log.Tracef("generator.GenerateServices called with gopt: %v", gopt)

	if len(gopt.Names) == 0 {
		gopt.Names = append(gopt.Names, "service-1")
	}

	gen.backupEnv()
	defer gen.restoreEnv()

	var files []string

	for _, service := range gopt.Names {

		gen.setEnv(map[string]string{
			"SERVICE_NAME":     service,
			"SERVICE_PRIORITY": strconv.Itoa(gopt.Priority),
		})

		sd := filepath.Join(gen.dockerfileServicesDir(), service)

		var rps []string
		if !gopt.Optional {
			rps = []string{".optional.template", "download.sh.template"}
		}

		// add service
		sf, err := gen.output("/services/service-name", sd, rps...)
		if err != nil {
			return nil, err
		}
		files = append(files, sf...)

		// add services extra files
		extraFiles := make([]string, 0, len(gopt.Tags))
		for _, tag := range gopt.Tags {
			extraFiles = append(extraFiles, filepath.Join(sd, gen.svcs.Config().TagsDir, tag))
		}

		ef, err := gen.create(extraFiles...)
		if err != nil {
			return nil, err
		}
		files = append(files, ef...)
	}

	return files, nil
}

func (gen *generator) Config() *GeneratorConfig {
	return gen.config
}

func (gen *generator) output(path string, dest string, excls ...string) ([]string, error) {

	log.Tracef("generator.output called with path: %v, dest: %v", path, dest)

	tmpDir, err := os.MkdirTemp(gen.fs.Paths().RunRoot, "gen-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	if err := helpers.Copy(filepath.Join(gen.fs.Paths().GeneratorTemplates, path), filepath.Join(tmpDir, dest)); err != nil {
		return nil, err
	}

	for _, excl := range excls {
		if err := helpers.Remove(filepath.Join(tmpDir, dest, excl)); err != nil {
			return nil, err
		}
	}

	return helpers.EnvsubstTemplates(tmpDir, gen.fs.Paths().RunGenerator, ".template")
}

func (gen *generator) create(paths ...string) ([]string, error) {

	log.Tracef("generator.create called with paths: %v", paths)

	files := make([]string, 0, len(paths))
	for _, p := range paths {

		dest := filepath.Join(gen.fs.Paths().RunGenerator, p)

		if _, err := helpers.Create(dest); err != nil {
			return nil, err
		}

		files = append(files, dest)
	}

	return files, nil
}

func (gen *generator) backupEnv() {
	gen.envBackup = os.Environ()
}

func (gen *generator) restoreEnv() {
	os.Clearenv()

	for _, e := range gen.envBackup {
		kv := strings.Split(e, "=")
		os.Setenv(kv[0], kv[1])
	}
}

func (gen *generator) setEnv(kv map[string]string) {
	os.Clearenv()

	for k, v := range kv {
		os.Setenv(k, v)
	}
}

func (gen *generator) envDir(child int) (string, error) {

	log.Tracef("generator.envDir called with child: %v", child)

	envDir := gen.fs.Paths().Environment

	if child != 0 {
		envDir = filepath.Join(envDir, fmt.Sprintf("%v-child", child))
	}

	envFiles, err := os.ReadDir(envDir)
	if goerrors.Is(err, os.ErrNotExist) {
		return envDir, nil
	}

	if len(envFiles) != 0 {
		return gen.envDir(child + 1)
	}

	return envDir, nil
}

func (gen *generator) dockerfileEnvDir() string {
	return filepath.Base(gen.fs.Paths().Environment)
}

func (gen *generator) dockerfileServicesDir() string {
	return filepath.Base(gen.fs.Paths().Services)
}
