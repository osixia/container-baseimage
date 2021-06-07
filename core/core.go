package core

import (
	"context"
	"embed"
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/osixia/container-baseimage/log"
)

// Core instance
var ci Core

// Core instance singleton creation lock
var ciLock = &sync.Mutex{}

// Core configuration
type CoreConfig struct {

	// Container image name
	Image string

	SupportedDistributions []*SupportedDistribution

	EnvironmentConfig *EnvironmentConfig
	FilesystemConfig  *FilesystemConfig
	ServicesConfig    *ServicesConfig
	GeneratorConfig   *GeneratorConfig
}

func (cc *CoreConfig) Validate() (bool, error) {

	if cc.Image == "" {
		return false, fmt.Errorf("%v: %w", "Image", ErrRequired)
	}

	if cc.SupportedDistributions == nil || len(cc.SupportedDistributions) == 0 {
		return false, fmt.Errorf("%v: %w", "SupportedDistributions", ErrRequired)
	}

	if cc.EnvironmentConfig == nil {
		return false, fmt.Errorf("%v: %w", "EnvironmentConfig", ErrRequired)
	}

	if cc.FilesystemConfig == nil {
		return false, fmt.Errorf("%v: %w", "FilesystemConfig", ErrRequired)
	}

	if cc.ServicesConfig == nil {
		return false, fmt.Errorf("%v: %w", "ServicesConfig", ErrRequired)
	}

	if cc.GeneratorConfig == nil {
		return false, fmt.Errorf("%v: %w", "GeneratorConfig", ErrRequired)
	}

	if _, err := cc.EnvironmentConfig.Validate(); err != nil {
		return false, fmt.Errorf("field %v is not valid: %w", "EnvironmentConfig", err)
	}

	if _, err := cc.FilesystemConfig.Validate(); err != nil {
		return false, fmt.Errorf("field %v is not valid: %w", "FilesystemConfig", err)
	}

	if _, err := cc.ServicesConfig.Validate(); err != nil {
		return false, fmt.Errorf("field %v is not valid: %w", "ServicesConfig", err)
	}

	if _, err := cc.GeneratorConfig.Validate(); err != nil {
		return false, fmt.Errorf("field %v is not valid: %w", "GeneratorConfig", err)
	}

	return true, nil
}

type core struct {
	env  Environment
	dist Distribution
	fs   Filesystem
	svcs Services
	gen  Generator
	ep   Entrypoint

	config *CoreConfig
}

type Core interface {
	Environment() Environment
	Distribution() Distribution
	Filesystem() Filesystem
	Services() Services
	Generator() Generator
	Entrypoint() Entrypoint

	Install(ctx context.Context) error

	Config() *CoreConfig
}

func Init(cc *CoreConfig) error {

	ciLock.Lock()
	defer ciLock.Unlock()

	// get core environment variables
	env, err := NewEnvironment(cc.EnvironmentConfig)
	if err != nil {
		return err
	}

	// customize core configuration with environment variables
	if env.ImageName() != "" {
		image := env.ImageName()

		if env.ImageTag() != "" {
			image = fmt.Sprintf("%v:%v", env.ImageName(), env.ImageTag())
		}

		log.Tracef("Setting container image from environment variables to %v ...", image)
		cc.Image = image
	}

	if env.DebugPackages() != "" {
		log.Tracef("Adding %v debug packages from environment variables ...", env.DebugPackages())

		esd := &SupportedDistribution{
			Name:    "Environment based distributions common configuration",
			Vendors: nil, // all vendors

			Config: &DistributionConfig{
				DebugPackages: strings.Split(env.DebugPackages(), " "),
			},
		}

		cc.SupportedDistributions = append(cc.SupportedDistributions, esd)
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
		log.Fatal("Core not initialized")
	}

	return ci
}

func (c *core) Config() *CoreConfig {
	return c.config
}

func (c *core) Environment() Environment {
	return c.env
}

func (c *core) Distribution() Distribution {

	if c.dist == nil {

		fs := c.Filesystem()
		svcs := c.Services()

		ciLock.Lock()
		defer ciLock.Unlock()

		dist, err := NewDistribution(fs, svcs, c.config.SupportedDistributions)
		if err != nil {
			log.Fatalf("Core: %v", err.Error())
		}
		c.dist = dist
	}

	return c.dist
}

func (c *core) Filesystem() Filesystem {

	if c.fs == nil {

		ciLock.Lock()
		defer ciLock.Unlock()

		fs, err := NewFilesystem(c.config.FilesystemConfig)
		if err != nil {
			log.Fatalf("Core: %v", err.Error())
		}
		c.fs = fs
	}

	return c.fs
}

func (c *core) Services() Services {

	if c.svcs == nil {

		fs := c.Filesystem()

		ciLock.Lock()
		defer ciLock.Unlock()

		svsc, err := NewServices(fs, c.config.ServicesConfig)
		if err != nil {
			log.Fatalf("Core: %v", err.Error())
		}
		c.svcs = svsc
	}

	return c.svcs
}

func (c *core) Generator() Generator {

	if c.gen == nil {

		env := c.Environment()
		fs := c.Filesystem()
		svcs := c.Services()
		ep := c.Entrypoint()
		dist := c.Distribution()

		ciLock.Lock()
		defer ciLock.Unlock()

		c.config.GeneratorConfig.fromImage = c.config.Image

		gen, err := NewGenerator(env, fs, svcs, ep, dist, c.config.GeneratorConfig)
		if err != nil {
			log.Fatalf("Core: %v", err.Error())
		}
		c.gen = gen
	}

	return c.gen
}

func (c *core) Entrypoint() Entrypoint {

	if c.ep == nil {

		fs := c.Filesystem()
		svcs := c.Services()
		dist := c.Distribution()

		ciLock.Lock()
		defer ciLock.Unlock()

		ep, err := NewEntrypoint(fs, dist, svcs)
		if err != nil {
			log.Fatalf("Core: %v", err.Error())
		}
		c.ep = ep
	}

	return c.ep
}

func (c *core) Install(ctx context.Context) error {

	log.Trace("distribution.Install called")

	log.Info("Creating container filesystem ...")
	if err := c.Filesystem().Create(); err != nil {
		return err
	}

	log.Infof("Copying %v assets to container filesystem ...", c.Distribution().Name())
	for _, assets := range c.Distribution().Config().Assets {
		if err := c.copyAssets(assets); err != nil {
			return err
		}
	}

	log.Infof("Linking %v files to %v ...", c.Filesystem().Paths().Bin, c.Distribution().Config().BinDest)
	if err := SymlinkAll(c.Filesystem().Paths().Bin, c.Distribution().Config().BinDest); err != nil {
		return err
	}

	// exec container install.sh script
	installSh := filepath.Join(c.Filesystem().Paths().Root, c.Distribution().Config().InstallScript)

	subCtx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	exec := NewExec(subCtx)

	if err := exec.Script(installSh); err != nil {
		return err
	}

	// remove container install.sh script
	log.Infof("Removing %v ...", installSh)
	if err := os.Remove(installSh); err != nil {
		return err
	}

	return nil
}

func (c *core) copyAssets(efs *embed.FS) error {

	log.Tracef("distribution.copyAssets called with efs: %+v", efs)

	if err := CopyEmbedDir(efs, c.Filesystem().Paths().Root, c.assetPerm); err != nil {
		return err
	}

	return nil
}

func (c *core) assetPerm(file string) iofs.FileMode {

	log.Tracef("distribution.assetPerm called with file: %v", file)

	var perm iofs.FileMode = 0644

	// add execute to files in container bin directory and .sh files
	if strings.HasPrefix(file, c.Filesystem().Paths().Bin) || strings.HasSuffix(file, ".sh") || strings.HasSuffix(file, ".sh"+c.Generator().Config().TemplatesFilesSuffix) {
		perm = 0755
	}

	return perm
}
