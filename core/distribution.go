package core

import (
	"bufio"
	"context"
	"embed"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/osixia/container-baseimage/log"
)

type distribution struct {
	fs   Filesystem
	svcs Services

	name            string
	vendor          string
	version         string
	versionCodename string // can be empty

	config *DistributionConfig
}

type Distribution interface {
	Name() string
	Vendor() string
	Version() string
	VersionCodename() string

	InstallPackages(ctx context.Context, packages []string) error
	AddMultiprocessStack(ctx context.Context) error

	Config() *DistributionConfig
}

type SupportedDistribution struct {
	Name    string
	Vendors []string

	Config *DistributionConfig
}

type DistributionConfig struct {
	MultiprocessStackServices []string
	DebugPackages             []string

	Assets []*embed.FS

	InstallScript string

	BinDest string

	BinAddMultiprocessStack     string
	BinPackagesIndexUpdate      string
	BinPackagesInstallClean     string
	BinPackagesIndexClean       string
	BinServicesInstall          string
	BinServicesLinkToEntrypoint string
}

func (dc *DistributionConfig) Merge(mdc *DistributionConfig) {

	if mdc.MultiprocessStackServices != nil {
		dc.MultiprocessStackServices = append(dc.MultiprocessStackServices, mdc.MultiprocessStackServices...)
	}
	if mdc.DebugPackages != nil {
		dc.DebugPackages = append(dc.DebugPackages, mdc.DebugPackages...)
	}

	if mdc.Assets != nil {
		dc.Assets = append(dc.Assets, mdc.Assets...)
	}

	if mdc.InstallScript != "" {
		dc.InstallScript = mdc.InstallScript
	}

	if mdc.BinDest != "" {
		dc.BinDest = mdc.BinDest
	}

	if mdc.BinAddMultiprocessStack != "" {
		dc.BinAddMultiprocessStack = mdc.BinAddMultiprocessStack
	}
	if mdc.BinPackagesIndexUpdate != "" {
		dc.BinPackagesIndexUpdate = mdc.BinPackagesIndexUpdate
	}
	if mdc.BinPackagesInstallClean != "" {
		dc.BinPackagesInstallClean = mdc.BinPackagesInstallClean
	}
	if mdc.BinPackagesIndexClean != "" {
		dc.BinPackagesIndexClean = mdc.BinPackagesIndexClean
	}
	if mdc.BinServicesInstall != "" {
		dc.BinServicesInstall = mdc.BinServicesInstall
	}
	if mdc.BinServicesLinkToEntrypoint != "" {
		dc.BinServicesLinkToEntrypoint = mdc.BinServicesLinkToEntrypoint
	}

}

func (dc *DistributionConfig) Validate() (bool, error) {

	if len(dc.MultiprocessStackServices) == 0 {
		return false, fmt.Errorf("%v: %w", "MultiprocessStackServices", ErrRequired)
	}

	if dc.InstallScript == "" {
		return false, fmt.Errorf("%v: %w", "InstallScript", ErrRequired)
	}

	if dc.BinDest == "" {
		return false, fmt.Errorf("%v: %w", "BinDest", ErrRequired)
	}

	if dc.BinAddMultiprocessStack == "" {
		return false, fmt.Errorf("%v: %w", "BinAddMultiprocessStack", ErrRequired)
	}
	if dc.BinPackagesIndexUpdate == "" {
		return false, fmt.Errorf("%v: %w", "BinPackagesIndexUpdate", ErrRequired)
	}
	if dc.BinPackagesInstallClean == "" {
		return false, fmt.Errorf("%v: %w", "BinPackagesInstallClean", ErrRequired)
	}
	if dc.BinPackagesIndexClean == "" {
		return false, fmt.Errorf("%v: %w", "BinPackagesIndexClean", ErrRequired)
	}
	if dc.BinServicesInstall == "" {
		return false, fmt.Errorf("%v: %w", "BinServicesInstall", ErrRequired)
	}
	if dc.BinServicesLinkToEntrypoint == "" {
		return false, fmt.Errorf("%v: %w", "BinServicesLinkToEntrypoint", ErrRequired)
	}

	return true, nil
}

func NewDistribution(fs Filesystem, svcs Services, sds []*SupportedDistribution) (Distribution, error) {

	f, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := &distribution{
		fs:   fs,
		svcs: svcs,
	}

	s := bufio.NewScanner(f)
	for s.Scan() {
		if m := regexp.MustCompile(`^PRETTY_NAME=(.*)$`).FindStringSubmatch(s.Text()); m != nil {
			d.name = strings.Trim(m[1], `"`)
		} else if m := regexp.MustCompile(`^ID=(.*)$`).FindStringSubmatch(s.Text()); m != nil {
			d.vendor = strings.Trim(m[1], `"`)
		} else if m := regexp.MustCompile(`^VERSION_ID=(.*)$`).FindStringSubmatch(s.Text()); m != nil {
			d.version = strings.Trim(m[1], `"`)
		} else if m := regexp.MustCompile(`^VERSION_CODENAME=(.*)$`).FindStringSubmatch(s.Text()); m != nil {
			d.versionCodename = strings.Trim(m[1], `"`)
		}
	}

	if d.name == "" || d.vendor == "" || d.version == "" {
		return nil, fmt.Errorf("%+v: %w", d, ErrDistributionNotFound)
	}

	d.config = &DistributionConfig{}

	vendor := strings.ToLower(d.vendor)
	for _, sd := range sds {
		log.Tracef("Supported distribution: %v", sd.Name)
		// nil vendors -> all vendors
		if sd.Vendors == nil {
			log.Tracef("Use \"%v\" config (apply to all vendors) ...", sd.Name)
			d.config.Merge(sd.Config)
			continue
		}

		// iterate distribution config vendors
		for _, distVendor := range sd.Vendors {
			if distVendor == vendor {
				log.Tracef("Use \"%v\" config (%v vendor match this config) ...", sd.Name, vendor)
				d.config.Merge(sd.Config)
			}
		}
	}

	if _, err := d.config.Validate(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *distribution) Name() string {
	return d.name
}

func (d *distribution) Vendor() string {
	return d.vendor
}

func (d *distribution) Version() string {
	return d.version
}

func (d *distribution) VersionCodename() string {
	return d.versionCodename
}

func (d *distribution) Config() *DistributionConfig {
	return d.config
}

func (d *distribution) InstallPackages(ctx context.Context, packages []string) error {

	log.Tracef("distribution.InstallPackages called with packages: %v", packages)

	if len(packages) == 0 {
		return nil
	}

	subCtx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	exec := NewExec(subCtx)

	scripts := []string{
		d.config.BinPackagesIndexUpdate,
		fmt.Sprintf("%v %v", d.config.BinPackagesInstallClean, strings.Join(packages, " ")),
	}

	if err := exec.Scripts(scripts); err != nil {
		return err
	}

	return nil
}

func (d *distribution) AddMultiprocessStack(ctx context.Context) error {

	log.Trace("distribution.AddMultiprocessStack called")

	multiprocessServices := d.config.MultiprocessStackServices

	if len(multiprocessServices) == 0 {
		log.Warningf("No service defined for the multiprocess stack.")
		return nil
	}

	log.Infof("Add multiprocess stack %v", strings.Join(multiprocessServices, ", "))

	services := make([]Service, 0, len(multiprocessServices))
	for _, name := range multiprocessServices {
		service, err := d.svcs.Get(name)
		if err != nil {
			return err
		}

		services = append(services, service)
	}

	if err := d.svcs.Require(ctx, services); err != nil {
		return err
	}

	return nil
}
