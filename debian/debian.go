package debian

import (
	"embed"

	"github.com/osixia/container-baseimage/core"
)

//go:embed assets/*
var assets embed.FS

var SupportedDistribution = &core.SupportedDistribution{
	Name:    "Debian & derivatives",
	Vendors: []string{"debian", "ubuntu"},

	Config: &core.DistributionConfig{
		DebugPackages: []string{"curl", "less", "procps", "psmisc", "strace", "vim-tiny"},
		Assets:        []*embed.FS{&assets},

		InstallScript: "debian.sh",

		InstallPackagesCmd: "apt-get install -y",
		UpdatePackagesCmd:  "apt-get -y update",
		CleanPackagesCmd:   "apt-get clean",
		RemovePackagesCmd:  "apt-get remove -y --purge --autoremove",
	},
}
