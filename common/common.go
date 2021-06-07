package common

import (
	"embed"

	"github.com/osixia/container-baseimage/core"
)

// list generator templates environment and service-name so .env and .priority files are included (. files are ignored in subdirs otherwise)

//go:embed assets/* assets/generator/templates/environment/* assets/generator/templates/services/service-name/*
var assets embed.FS

var CommonSupportedDistribution = &core.SupportedDistribution{
	Name:    "Distributions common configuration",
	Vendors: nil, // all vendors

	Config: &core.DistributionConfig{
		Assets: []*embed.FS{&assets},

		BinDest: "/usr/sbin",

		BinAddMultiprocessStack:     "add-multiprocess-stack",
		BinServicesInstall:          "services-install",
		BinServicesLinkToEntrypoint: "entrypoint-link-services",
	},
}
