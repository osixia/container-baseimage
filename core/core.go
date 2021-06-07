package core

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

// Core global variables
// =============================
var (
	CoreConfig *Config

	ci     Core      // core instance
	ciOnce sync.Once // core instance singleton creation guard
)

// Core functions
// =============================

func Init(cc *Config) error {

	log.Trace("core.Init called")

	if cc == nil {
		return fmt.Errorf("cc: %w", errors.ErrRequired)
	}

	// get core environment variables
	env, err := newEnvironment(cc.EnvironmentConfig)
	if err != nil {
		return err
	}

	// customize core configuration with environment variables
	if env.Image() != "" {
		image := env.Image()

		log.Tracef("Setting container image from environment variables to %v ...", image)
		cc.Image = image
	}

	if env.DebugPackages() != "" {
		log.Tracef("Adding %v debug packages from environment variables ...", env.DebugPackages())

		sd := &SupportedDistribution{
			Name:    "Environment based distributions common configuration",
			Vendors: nil, // all vendors

			Config: &DistributionConfig{
				DebugPackages: strings.Split(env.DebugPackages(), " "),
			},
		}

		cc.SupportedDistributions = append(cc.SupportedDistributions, sd)
	}

	// validate config
	if _, err := cc.Validate(); err != nil {
		return err
	}

	ci = &core{
		env:    env,
		config: cc,
	}

	return nil
}

func Instance() Core {

	ciOnce.Do(func() {
		helpers.Mustf(Init(CoreConfig), "Error initializing core")
	})

	return ci
}

// Core config
// =============================

type Config struct {

	// Container image name
	Image string

	SupportedDistributions []*SupportedDistribution

	EnvironmentConfig *EnvironmentConfig
	FilesystemConfig  *FilesystemConfig
	ServicesConfig    *ServicesConfig
	ProcessesConfig   *ProcessesConfig
}

func (cc *Config) Validate() (bool, error) {

	if cc.Image == "" {
		return false, fmt.Errorf("Image: %w", errors.ErrRequired)
	}

	if len(cc.SupportedDistributions) == 0 {
		return false, fmt.Errorf("SupportedDistributions: %w", errors.ErrRequired)
	}

	if cc.EnvironmentConfig == nil {
		return false, fmt.Errorf("EnvironmentConfig: %w", errors.ErrRequired)
	}
	if cc.FilesystemConfig == nil {
		return false, fmt.Errorf("FilesystemConfig: %w", errors.ErrRequired)
	}
	if cc.ServicesConfig == nil {
		return false, fmt.Errorf("ServicesConfig: %w", errors.ErrRequired)
	}
	if cc.ProcessesConfig == nil {
		return false, fmt.Errorf("ProcessesConfig: %w", errors.ErrRequired)
	}

	return true, nil
}

// Core
// =============================

type Core interface {
	Install(ctx context.Context) error

	Environment() Environment
	Distribution() Distribution
	Filesystem() Filesystem
	Services() Services
	Processes() Processes
	Entrypoint() Entrypoint
	Generator() Generator

	Config() *Config
}

type core struct {
	env Environment

	dist     Distribution
	distOnce sync.Once

	fs     Filesystem
	fsOnce sync.Once

	svcs     Services
	svcsOnce sync.Once

	prcs     Processes
	prcsOnce sync.Once

	ep     Entrypoint
	epOnce sync.Once

	gen     Generator
	genOnce sync.Once

	config *Config
}

func (c *core) Install(ctx context.Context) error {

	log.Trace("core.Install called")

	if err := c.Filesystem().Create(); err != nil {
		return err
	}

	di := newDistributionInstaller(c.Filesystem(), c.Distribution())
	return di.Install(ctx)
}

func (c *core) Environment() Environment {
	return c.env
}

func (c *core) Distribution() Distribution {

	c.distOnce.Do(func() {
		c.dist = helpers.MustVal(newDistribution(c.config.SupportedDistributions))
	})

	return c.dist
}

func (c *core) Filesystem() Filesystem {

	c.fsOnce.Do(func() {
		c.fs = helpers.MustVal(newFilesystem(c.config.FilesystemConfig))
	})

	return c.fs
}

func (c *core) Services() Services {

	c.svcsOnce.Do(func() {
		fs := c.Filesystem()
		c.svcs = helpers.MustVal(newServices(fs, c.config.ServicesConfig))
	})

	return c.svcs
}

func (c *core) Processes() Processes {

	c.prcsOnce.Do(func() {
		fs := c.Filesystem()
		c.prcs = helpers.MustVal(newProcesses(fs, c.config.ProcessesConfig))
	})

	return c.prcs
}

func (c *core) Entrypoint() Entrypoint {

	c.epOnce.Do(func() {

		fs := c.Filesystem()
		svcs := c.Services()
		prcs := c.Processes()

		c.ep = helpers.MustVal(newEntrypoint(fs, svcs, prcs))
	})

	return c.ep
}

func (c *core) Generator() Generator {

	c.genOnce.Do(func() {

		env := c.Environment()
		fs := c.Filesystem()
		svcs := c.Services()

		genc := &GeneratorConfig{
			fromImage: c.config.Image,
		}

		c.gen = helpers.MustVal(newGenerator(env, fs, svcs, genc))
	})

	return c.gen
}

func (c *core) Config() *Config {
	return c.config
}
