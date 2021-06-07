package config

import (
	"fmt"

	"github.com/osixia/container-baseimage/alpine"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/debian"
	"github.com/osixia/container-baseimage/log"
)

// global variables
var (
	ImageName    = "osixia/baseimage"
	ImageTag     = "develop"
	Version      = "develop"
	Contributors = "🐒✨🌴"
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

// core supported distributions
var CoreSupportedDistributions = []*core.SupportedDistribution{
	alpine.SupportedDistribution,
	debian.SupportedDistribution,
}

// core filesystem configuration
var CoreFilesystemConfig = &core.FilesystemConfig{
	RootPath:    "/container",
	RunRootPath: "/run/container",

	EnvironmentFilesPrefix: ".env",
}

// core services configuration
var CoreServicesConfig = &core.ServicesConfig{
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

// core processes configuration
var CoreProcessesConfig = &core.ProcessesConfig{
	PIDFileSuffix:        ".pid",
	WantedDownFileSuffix: ".down",
	TagsDir:              ".tags",
}

// core configuration
var CoreConfig = &core.Config{
	Image:                  fmt.Sprintf("%v:%v", ImageName, ImageTag),
	SupportedDistributions: CoreSupportedDistributions,

	EnvironmentConfig: CoreEnvironmentConfig,
	FilesystemConfig:  CoreFilesystemConfig,
	ServicesConfig:    CoreServicesConfig,
	ProcessesConfig:   CoreProcessesConfig,
}
