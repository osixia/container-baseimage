package core

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/log"
)

// Core global variables
// =============================
var (
	CoreConfig *Config

	ci     Core            // core instance
	ciLock = &sync.Mutex{} // core instance singleton creation lock
)

// Core functions
// =============================

func Init(cc *Config) error {

	log.Trace("core.Init called")

	if cc == nil {
		return fmt.Errorf("cc: %w", errors.ErrRequired)
	}

	ciLock.Lock()
	defer ciLock.Unlock()

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
	if ci == nil {
		if err := Init(CoreConfig); err != nil {
			log.Fatalf("Error initializing core: %v", err.Error())
		}
	}

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
	env  Environment
	dist Distribution
	fs   Filesystem
	svcs Services
	prcs Processes
	gen  Generator
	ep   Entrypoint

	config *Config
}

func (c *core) Install(ctx context.Context) error {

	log.Trace("core.Install called")

	if err := c.Filesystem().Create(); err != nil {
		return err
	}

	di := newDistributionInstaller(c.Filesystem(), c.Distribution())
	if err := di.Install(ctx); err != nil {
		return err
	}

	return nil
}

func (c *core) Environment() Environment {
	return c.env
}

func (c *core) Distribution() Distribution {

	if c.dist == nil {

		ciLock.Lock()
		defer ciLock.Unlock()

		var err error

		c.dist, err = newDistribution(c.config.SupportedDistributions)
		if err != nil {
			log.Fatalf("Core: %v", err.Error())
		}
	}

	return c.dist
}

func (c *core) Filesystem() Filesystem {

	if c.fs == nil {

		ciLock.Lock()
		defer ciLock.Unlock()

		var err error

		c.fs, err = newFilesystem(c.config.FilesystemConfig)
		if err != nil {
			log.Fatalf("Core: %v", err.Error())
		}
	}

	return c.fs
}

func (c *core) Services() Services {

	if c.svcs == nil {

		fs := c.Filesystem()

		ciLock.Lock()
		defer ciLock.Unlock()

		var err error

		c.svcs, err = newServices(fs, c.config.ServicesConfig)
		if err != nil {
			log.Fatalf("Core: %v", err.Error())
		}
	}

	return c.svcs
}

func (c *core) Processes() Processes {

	if c.prcs == nil {

		fs := c.Filesystem()

		ciLock.Lock()
		defer ciLock.Unlock()

		var err error

		c.prcs, err = newProcesses(fs, c.config.ProcessesConfig)
		if err != nil {
			log.Fatalf("Core: %v", err.Error())
		}
	}

	return c.prcs
}

func (c *core) Entrypoint() Entrypoint {

	if c.ep == nil {

		fs := c.Filesystem()
		dist := c.Distribution()
		svcs := c.Services()
		prcs := c.Processes()

		ciLock.Lock()
		defer ciLock.Unlock()

		var err error

		c.ep, err = newEntrypoint(fs, dist, svcs, prcs)
		if err != nil {
			log.Fatalf("Core: %v", err.Error())
		}
	}

	return c.ep
}

func (c *core) Generator() Generator {

	if c.gen == nil {

		env := c.Environment()
		fs := c.Filesystem()
		svcs := c.Services()

		genc := &GeneratorConfig{
			fromImage: c.config.Image,
		}

		ciLock.Lock()
		defer ciLock.Unlock()

		var err error

		c.gen, err = newGenerator(env, fs, svcs, genc)
		if err != nil {
			log.Fatalf("Core: %v", err.Error())
		}
	}

	return c.gen
}

func (c *core) Config() *Config {
	return c.config
}
