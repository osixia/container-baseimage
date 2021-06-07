package config

import (
	"fmt"

	"github.com/osixia/container-baseimage/alpine"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/debian"
	"github.com/osixia/container-baseimage/log"
)

// build global variables
var (
	BuildVersion      = "develop"
	BuildContributors = "🐒✨🌴"

	BuildImageName = "osixia/baseimage"
	BuildImageTag  = "develop"
)

// global variables
var (
	EnvsubstTemplatesFilesSuffix = ".template"
	TagsNamePrefix               = "tag:"
)

// logger environment configuration
var LogEnvironmentConfig = &log.EnvironmentConfig{
	LevelKey:  "CONTAINER_LOG_LEVEL",
	FormatKey: "CONTAINER_LOG_FORMAT",
}

// core environment configuration
var CoreEnvironmentConfig = &core.EnvironmentConfig{
	ImageKey:         "CONTAINER_IMAGE",
	DebugPackagesKey: "CONTAINER_DEBUG_PACKAGES",
}

// supported distributions
var SupportedDistributions = []*core.SupportedDistribution{
	alpine.SupportedDistribution,
	debian.SupportedDistribution,
}

// filesystem configuration
var FilesystemConfig = &core.FilesystemConfig{
	RootPath:    "/container",
	RunRootPath: "/run/container",

	EnvironmentFilesPrefix: ".env",
}

// services configuration
var ServicesConfig = &core.ServicesConfig{
	TagsDir: ".tags",

	DefaultPriority:  500,
	PriorityFilename: ".priority",

	OptionalFilename: ".optional",

	DownloadFilename:   "download.sh",
	DownloadedFilename: ".downloaded",

	InstallFilename:   "install.sh",
	InstalledFilename: ".installed",

	StartupFilename: "startup.sh",
	ProcessFilename: "process.sh",
	FinishFilename:  "finish.sh",

	LinkedFilename: ".entrypoint",
}

// processes configuration
var ProcessesConfig = &core.ProcessesConfig{
	PIDFileSuffix:        ".pid",
	WantedDownFileSuffix: ".down",
	TagsDir:              ".tags",
}

// core configuration
var CoreConfig = &core.CoreConfig{
	Image:                  fmt.Sprintf("%v:%v", BuildImageName, BuildImageTag),
	SupportedDistributions: SupportedDistributions,

	EnvironmentConfig: CoreEnvironmentConfig,
	FilesystemConfig:  FilesystemConfig,
	ServicesConfig:    ServicesConfig,
	ProcessesConfig:   ProcessesConfig,
}
