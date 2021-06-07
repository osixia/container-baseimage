package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/osixia/container-baseimage/log"
)

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
	ep   Entrypoint
	dist Distribution

	envBackup []string

	config *GeneratorConfig
}

type GenerateBootstrapOptions struct {
	GenerateDockerfileOptions
	GenerateServicesOptions
}

type GenerateDockerfileOptions struct {
	Multiprocess bool
}

type GenerateServicesOptions struct {
	Names    []string
	Priority int
	Optional bool
}

type GeneratorConfig struct {
	fromImage string

	TemplatesFilesSuffix string

	DockerfileTemplate             string
	DockerfileMultiprocessTemplate string

	ServicesTemplatesDir string

	EnvironmentTemplatesDir string
}

func (genc *GeneratorConfig) Validate() (bool, error) {

	if genc.TemplatesFilesSuffix == "" {
		return false, fmt.Errorf("%v: %w", "TemplatesFilesSuffix", ErrRequired)
	}

	if genc.DockerfileTemplate == "" {
		return false, fmt.Errorf("%v: %w", "DockerfileTemplate", ErrRequired)
	}
	if genc.DockerfileMultiprocessTemplate == "" {
		return false, fmt.Errorf("%v: %w", "DockerfileMultiprocessTemplate", ErrRequired)
	}

	if genc.ServicesTemplatesDir == "" {
		return false, fmt.Errorf("%v: %w", "ServicesTemplatesDir", ErrRequired)
	}

	if genc.EnvironmentTemplatesDir == "" {
		return false, fmt.Errorf("%v: %w", "EnvironmentTemplatesDir", ErrRequired)
	}

	return true, nil
}

func NewGenerator(env Environment, fs Filesystem, svcs Services, ep Entrypoint, dist Distribution, genc *GeneratorConfig) (Generator, error) {

	if _, err := genc.Validate(); err != nil {
		return nil, err
	}

	return &generator{
		env:    env,
		fs:     fs,
		svcs:   svcs,
		ep:     ep,
		dist:   dist,
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

	files := []string{}

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

	log.Tracef("generator.GenerateDockerfile called with gopt: %v", gopt)

	dockerfile := gen.config.DockerfileTemplate

	processes, err := gen.ep.ListLinkedServices(LinkedServicesWithStep(LifecycleStepProcess))
	if err != nil {
		return nil, err
	}

	nbProcesses := len(processes)
	if nbProcesses == 1 || (gopt.Multiprocess && nbProcesses == 0) {
		dockerfile = gen.config.DockerfileMultiprocessTemplate
	}

	envDir, err := gen.envDir(0)
	if err != nil {
		return nil, err
	}

	gen.backupEnv()
	defer gen.restoreEnv()

	gen.setEnv(map[string]string{
		"FROM_IMAGE":                       gen.config.fromImage,
		"CONTAINER_IMAGE_NAME_ENV_KEY":     gen.env.Config().ImageNameKey,
		"CONTAINER_IMAGE_TAG_ENV_KEY":      gen.env.Config().ImageTagKey,
		"PACKAGES_INDEX_UPDATE_BIN":        gen.dist.Config().BinPackagesIndexUpdate,
		"ADD_MULTIPROCESS_STACK_BIN":       gen.dist.Config().BinAddMultiprocessStack,
		"PACKAGES_INSTALL_CLEAN_BIN":       gen.dist.Config().BinPackagesInstallClean,
		"DOCKERFILE_SERVICES_DIR":          gen.dockerfileServicesDir(),
		"INSTALL_SERVICES_BIN":             gen.dist.Config().BinServicesInstall,
		"LINK_SERVICES_ENTRYPOINT_BIN":     gen.dist.Config().BinServicesLinkToEntrypoint,
		"CONTAINER_SERVICES_DIR":           gen.fs.Paths().Services,
		"DOCKERFILE_ENVIRONMENT_FILES_DIR": gen.dockerfileEnvDir(),
		"CONTAINER_ENVIRONMENT_FILES_DIR":  envDir,
	})

	// add dockerfile
	return gen.output(dockerfile, "Dockerfile"+gen.config.TemplatesFilesSuffix)
}

func (gen *generator) envDir(child int) (string, error) {

	envDir := gen.fs.Paths().EnvironmentFiles

	if child != 0 {
		envDir = filepath.Join(envDir, fmt.Sprintf("%v-child", child))
	}

	envFiles, err := os.ReadDir(envDir)
	if errors.Is(err, os.ErrNotExist) {
		return envDir, nil
	}

	if len(envFiles) != 0 {
		return gen.envDir(child + 1)
	}

	return envDir, nil
}

func (gen *generator) GenerateEnvironment() ([]string, error) {

	log.Tracef("generator.GenerateEnvironment called")

	gen.backupEnv()
	defer gen.restoreEnv()

	gen.setEnv(nil)

	// add environment
	return gen.output(gen.config.EnvironmentTemplatesDir, gen.dockerfileEnvDir())
}

func (gen *generator) dockerfileEnvDir() string {
	return filepath.Base(gen.fs.Paths().EnvironmentFiles)
}

func (gen *generator) dockerfileServicesDir() string {
	return filepath.Base(gen.fs.Paths().Services)
}

func (gen *generator) GenerateServices(gopt *GenerateServicesOptions) ([]string, error) {

	log.Tracef("generator.GenerateServices called with gopt: %v", gopt)

	if len(gopt.Names) == 0 {
		gopt.Names = append(gopt.Names, "service-1")
	}

	gen.backupEnv()
	defer gen.restoreEnv()

	files := []string{}

	for _, service := range gopt.Names {

		gen.setEnv(map[string]string{
			"SERVICE_NAME":           service,
			"SERVICE_PRIORITY":       strconv.Itoa(gopt.Priority),
			"CONTAINER_RUN_ROOT_DIR": gen.fs.Paths().RunRoot,
		})

		// add service
		sf, err := gen.output(gen.config.ServicesTemplatesDir, filepath.Join(gen.dockerfileServicesDir(), service))
		if err != nil {
			return nil, err
		}

		// remove .optional and download.sh files for non-optional service
		if !gopt.Optional {
			for i, f := range sf {
				bf := filepath.Base(f)
				if bf == gen.svcs.Config().OptionalFilename || bf == gen.svcs.Config().DownloadFilename {
					if err := os.Remove(f); err != nil {
						return nil, err
					}
					sf = append(sf[:i], sf[i+1:]...)
				}
			}
		}

		files = append(files, sf...)
	}

	return files, nil
}

func (gen *generator) output(path string, dest string) ([]string, error) {

	log.Tracef("generator.output called with path: %v, dest: %v", path, dest)

	tmpDir, err := os.MkdirTemp(gen.fs.Paths().RunRoot, "gen-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	if err := Copy(filepath.Join(gen.fs.Paths().GeneratorTemplates, path), filepath.Join(tmpDir, dest)); err != nil {
		return nil, err
	}

	return EnvsubstTemplates(tmpDir, gen.fs.Paths().RunGeneratorOutput, gen.config.TemplatesFilesSuffix)
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

func (gen *generator) Config() *GeneratorConfig {
	return gen.config
}
