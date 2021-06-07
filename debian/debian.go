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

		InstallPackagesCmds: []string{"apt-get install -y --no-install-recommends --no-install-suggests"},
		UpdatePackagesCmds:  []string{"apt-get -y update"},
		CleanPackagesCmds:   []string{"apt-get clean", "find /var/lib/apt/lists -mindepth 1 -delete"},
		RemovePackagesCmds:  []string{"apt-get remove -y --purge --autoremove"},

		AddGroupCmdTpl: `groupadd -g {{.NonrootGroup.ID}} {{.NonrootGroup.Name}}`,
		AddUserCmdTpl:  `useradd -m -u {{.NonrootUser.ID}} -g {{.NonrootGroup.ID}} -s /sbin/nologin {{.NonrootUser.Name}}`,
	},
}
