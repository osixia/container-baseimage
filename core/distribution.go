package core

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"github.com/google/shlex"

	"github.com/osixia/container-baseimage/errors"
	"github.com/osixia/container-baseimage/helpers"
	"github.com/osixia/container-baseimage/log"
)

// list generator templates environment and service-name so .env and .priority files are included (dot files are ignored in subdirs otherwise)

//go:embed assets/* assets/generator/templates/environment/* assets/generator/templates/services/service-name/*
var assets embed.FS

// Core distribution
// =============================

func newCoreDistributionConfig() *DistributionConfig {
	return &DistributionConfig{
		Assets: []*embed.FS{&assets},
	}
}

// Supported distribution
// =============================

type SupportedDistribution struct {
	Name    string
	Vendors []string

	Config *DistributionConfig
}

// Distribution config
// =============================

type DistributionConfig struct {
	DebugPackages []string

	Assets []*embed.FS

	InstallScript string

	InstallPackagesCmds []string
	UpdatePackagesCmds  []string
	CleanPackagesCmds   []string
	RemovePackagesCmds  []string

	AddGroupCmdTpl string
	AddUserCmdTpl  string
}

func (dc *DistributionConfig) Merge(mdc *DistributionConfig) {

	if mdc.DebugPackages != nil {
		dc.DebugPackages = append(dc.DebugPackages, mdc.DebugPackages...)
	}

	if mdc.Assets != nil {
		dc.Assets = append(dc.Assets, mdc.Assets...)
	}

	if mdc.InstallScript != "" {
		dc.InstallScript = mdc.InstallScript
	}

	dc.InstallPackagesCmds = append(dc.InstallPackagesCmds, mdc.InstallPackagesCmds...)
	dc.UpdatePackagesCmds = append(dc.UpdatePackagesCmds, mdc.UpdatePackagesCmds...)
	dc.CleanPackagesCmds = append(dc.CleanPackagesCmds, mdc.CleanPackagesCmds...)
	dc.RemovePackagesCmds = append(dc.RemovePackagesCmds, mdc.RemovePackagesCmds...)

	if mdc.AddGroupCmdTpl != "" {
		dc.AddGroupCmdTpl = mdc.AddGroupCmdTpl
	}
	if mdc.AddUserCmdTpl != "" {
		dc.AddUserCmdTpl = mdc.AddUserCmdTpl
	}

}

func (dc *DistributionConfig) Validate() (bool, error) {

	if dc.InstallScript == "" {
		return false, fmt.Errorf("InstallScript: %w", errors.ErrRequired)
	}

	if len(dc.InstallPackagesCmds) == 0 {
		return false, fmt.Errorf("InstallPackagesCmds: %w", errors.ErrRequired)
	}
	if len(dc.UpdatePackagesCmds) == 0 {
		return false, fmt.Errorf("UpdatePackagesCmds: %w", errors.ErrRequired)
	}
	if len(dc.CleanPackagesCmds) == 0 {
		return false, fmt.Errorf("CleanPackagesCmds: %w", errors.ErrRequired)
	}
	if len(dc.RemovePackagesCmds) == 0 {
		return false, fmt.Errorf("RemovePackagesCmds: %w", errors.ErrRequired)
	}

	if dc.AddGroupCmdTpl == "" {
		return false, fmt.Errorf("AddGroupCmdTpl: %w", errors.ErrRequired)
	}
	if dc.AddUserCmdTpl == "" {
		return false, fmt.Errorf("AddUserCmdTpl: %w", errors.ErrRequired)
	}

	return true, nil
}

// Distribution
// =============================

type Distribution interface {
	Name() string
	Vendor() string
	Version() string
	VersionCodename() string

	InstallPackages(ctx context.Context, packages []string, update bool, clean bool) error
	UpdatePackages(ctx context.Context) error
	CleanPackages(ctx context.Context) error
	RemovePackages(ctx context.Context, packages []string) error

	AddGroup(ctx context.Context, id string, name string) error
	AddUser(ctx context.Context, id string, name string, groupID string, groupName string) error

	Config() *DistributionConfig
}

type distribution struct {
	name            string
	vendor          string
	version         string
	versionCodename string // can be empty

	config *DistributionConfig
}

func newDistribution(sds []*SupportedDistribution) (Distribution, error) {

	f, err := os.Open("/etc/os-release")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Warning(err.Error())
		}
	}()

	dist := &distribution{}

	s := bufio.NewScanner(f)

	nameRegexp := regexp.MustCompile(`^PRETTY_NAME=(.*)$`)
	vendorRegexp := regexp.MustCompile(`^ID=(.*)$`)
	versionRegexp := regexp.MustCompile(`^VERSION_ID=(.*)$`)
	versionCodenameRegexp := regexp.MustCompile(`^VERSION_CODENAME=(.*)$`)

	for s.Scan() {
		if m := nameRegexp.FindStringSubmatch(s.Text()); m != nil {
			dist.name = strings.Trim(m[1], `"`)
		} else if m := vendorRegexp.FindStringSubmatch(s.Text()); m != nil {
			dist.vendor = strings.Trim(m[1], `"`)
		} else if m := versionRegexp.FindStringSubmatch(s.Text()); m != nil {
			dist.version = strings.Trim(m[1], `"`)
		} else if m := versionCodenameRegexp.FindStringSubmatch(s.Text()); m != nil {
			dist.versionCodename = strings.Trim(m[1], `"`)
		}
	}

	if dist.name == "" || dist.vendor == "" || dist.version == "" {
		return nil, fmt.Errorf("%+v: distribution %w", dist, errors.ErrUnknown)
	}

	dist.config = newCoreDistributionConfig()

	vendor := strings.ToLower(dist.vendor)
	for _, sd := range sds {
		log.Tracef("Supported distribution: %v", sd.Name)

		if sd.Config == nil {
			continue
		}

		// nil vendors -> distribution configuration apply to all vendors
		if sd.Vendors == nil {
			dist.config.Merge(sd.Config)
			continue
		}

		// iterate distribution config vendors
		for _, distVendor := range sd.Vendors {
			if distVendor == vendor {
				log.Tracef("Use \"%v\" config (%v vendor match this config) ...", sd.Name, vendor)
				dist.config.Merge(sd.Config)
				break
			}
		}
	}

	if _, err := dist.config.Validate(); err != nil {
		return nil, err
	}

	return dist, nil
}

func (dist *distribution) Name() string {
	return dist.name
}

func (dist *distribution) Vendor() string {
	return dist.vendor
}

func (dist *distribution) Version() string {
	return dist.version
}

func (dist *distribution) VersionCodename() string {
	return dist.versionCodename
}

func (dist *distribution) InstallPackages(ctx context.Context, packages []string, update bool, clean bool) error {

	log.Tracef("distribution.InstallPackages called with packages: %v, update: %v, clean: %v", packages, update, clean)

	if len(packages) == 0 {
		return nil
	}

	subCtx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	if update {
		if err := dist.UpdatePackages(subCtx); err != nil {
			return err
		}
	}

	if err := dist.execPackagesCmds(ctx, dist.config.InstallPackagesCmds, packages...); err != nil {
		return err
	}

	if clean {
		if err := dist.CleanPackages(subCtx); err != nil {
			return err
		}
	}

	return nil
}

func (dist *distribution) UpdatePackages(ctx context.Context) error {

	log.Trace("distribution.UpdatePackages called")

	return dist.execPackagesCmds(ctx, dist.config.UpdatePackagesCmds)
}
func (dist *distribution) CleanPackages(ctx context.Context) error {

	log.Trace("distribution.CleanPackages called")

	return dist.execPackagesCmds(ctx, dist.config.CleanPackagesCmds)
}

func (dist *distribution) RemovePackages(ctx context.Context, packages []string) error {

	log.Tracef("distribution.RemovePackages called with packages: %v", packages)

	return dist.execPackagesCmds(ctx, dist.config.RemovePackagesCmds, packages...)
}

func (dist *distribution) AddGroup(ctx context.Context, id string, name string) error {

	log.Tracef("distribution.AddGroup called with id: %v, name: %v", id, name)

	data := map[string]any{
		"NonrootGroup": map[string]any{
			"ID":   id,
			"Name": name,
		},
	}

	return dist.execCmdTpl(ctx, dist.config.AddGroupCmdTpl, data)
}

func (dist *distribution) AddUser(ctx context.Context, id string, name string, groupID string, groupName string) error {

	log.Tracef("distribution.AddUser called with id: %v, name: %v, groupID: %v, groupName: %v", id, name, groupID, groupName)

	data := map[string]any{
		"NonrootGroup": map[string]any{
			"ID":   groupID,
			"Name": groupName,
		},
		"NonrootUser": map[string]any{
			"ID":   id,
			"Name": name,
		},
	}

	return dist.execCmdTpl(ctx, dist.config.AddUserCmdTpl, data)
}

func (dist *distribution) Config() *DistributionConfig {
	return dist.config
}

func (dist *distribution) execPackagesCmd(ctx context.Context, cmd string, args ...string) error {

	log.Tracef("distribution.execPackagesCmd called with cmd: %v %v", cmd, args)

	parts, err := shlex.Split(cmd)
	if err != nil {
		return fmt.Errorf("failed to parse command %v: %w", cmd, err)
	}

	allArgs := append(parts[1:], args...)

	exec := helpers.NewExec(ctx).WithStdoutLogPrintFunc(log.Info).WithStderrLogPrintFunc(log.Error)

	if slices.Contains([]string{"debian", "ubuntu"}, dist.Vendor()) {
		exec.WithEnv(append(os.Environ(), "DEBIAN_FRONTEND=noninteractive"))
	}

	return exec.Command(parts[0], allArgs...).Run()
}

func (dist *distribution) execPackagesCmds(ctx context.Context, cmds []string, args ...string) error {

	log.Tracef("distribution.execPackagesCmds called with cmds: %v %v", cmds, args)

	for _, cmd := range cmds {
		if err := dist.execPackagesCmd(ctx, cmd, args...); err != nil {
			return err
		}
	}

	return nil
}

func (dist *distribution) execCmdTpl(ctx context.Context, tpl string, data map[string]any) error {

	log.Tracef("distribution.execCmdTpl called with tpl: %v, data: %v", tpl, data)

	t, err := template.New("cmd").Parse(tpl)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = t.Execute(&out, data)
	if err != nil {
		return err
	}

	cmd := out.String()
	parts, err := shlex.Split(cmd)
	if err != nil {
		return fmt.Errorf("failed to parse command %v: %w", cmd, err)
	}

	exec := helpers.NewExec(ctx).WithStdoutLogPrintFunc(log.Info).WithStderrLogPrintFunc(log.Error)

	return exec.Command(parts[0], parts[1:]...).Run()
}

type DistributionInstaller interface {
	Install(ctx context.Context) error
}

type distributionInstaller struct {
	dist Distribution
	fs   Filesystem
}

func newDistributionInstaller(fs Filesystem, dist Distribution) DistributionInstaller {
	return &distributionInstaller{
		dist: dist,
		fs:   fs,
	}
}

func (di *distributionInstaller) Install(ctx context.Context) error {

	log.Trace("distributionInstaller.Install called")

	log.Infof("Copying %v assets to container filesystem ...", di.dist.Name())
	for _, assets := range di.dist.Config().Assets {
		if err := di.copyAssets(assets); err != nil {
			return err
		}
	}

	// exec distribution install script
	installSh := filepath.Join(di.fs.Paths().Root, di.dist.Config().InstallScript)

	subCtx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	if err := helpers.NewExec(subCtx).Command(installSh).Run(); err != nil {
		return err
	}

	// remove distribution install script
	return helpers.Remove(installSh)
}

func (di *distributionInstaller) copyAssets(efs *embed.FS) error {

	log.Tracef("distributionInstaller.copyAssets called with efs: %+v", efs)

	return helpers.CopyEmbedDir(efs, di.fs.Paths().Root, di.assetPerm)
}

func (di *distributionInstaller) assetPerm(file string) iofs.FileMode {

	log.Tracef("distributionInstaller.assetPerm called with file: %v", file)

	var perm iofs.FileMode = 0644

	// add execute to *.sh, *.sh.* files
	if strings.HasSuffix(file, ".sh") || strings.Contains(file, ".sh.") {
		perm = 0755
	}

	return perm
}
