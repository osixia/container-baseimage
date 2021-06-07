package alpine

import (
	"embed"

	"github.com/osixia/container-baseimage/core"
)

//go:embed assets/*
var assets embed.FS

var SupportedDistribution = &core.SupportedDistribution{
	Name:    "Alpine",
	Vendors: []string{"alpine"},

	Config: &core.DistributionConfig{
		DebugPackages: []string{"curl", "less", "procps", "psmisc", "strace", "vim"},
		Assets:        []*embed.FS{&assets},

		InstallScript: "alpine.sh",

		InstallPackagesCmd: "apk add",
		UpdatePackagesCmd:  "apk update",
		CleanPackagesCmd:   "apk cache purge",
		RemovePackagesCmd:  "apk del --purge",

		AddGroupCmdTpl: `addgroup -g {{.NonrootGroup.ID}} {{.NonrootGroup.Name}}`,
		AddUserCmdTpl:  `adduser -D -u {{.NonrootUser.ID}} -G {{.NonrootGroup.Name}} -s /sbin/nologin {{.NonrootUser.Name}}`,
	},
}
