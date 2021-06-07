package config

import (
	"fmt"

	"github.com/osixia/container-baseimage/alpine"
	"github.com/osixia/container-baseimage/common"
	"github.com/osixia/container-baseimage/core"
	"github.com/osixia/container-baseimage/debian"
	"github.com/osixia/container-baseimage/log"
)

// build variables
var BuildVersion = "develop"
var BuildContributors = "🐒✨🌴"

var BuildImageName = "osixia/baseimage"
var BuildImageTag = "develop"

// global variables
var EnvsubstTemplatesFilesSuffix = ".template"

// logger environment configuration
var LogEnvironmentConfig = &log.EnvironmentConfig{
	LevelKey:  "CONTAINER_LOG_LEVEL",
	FormatKey: "CONTAINER_LOG_FORMAT",
}

// core environment configuration
var CoreEnvironmentConfig = &core.EnvironmentConfig{
	ImageNameKey: "CONTAINER_IMAGE_NAME",
	ImageTagKey:  "CONTAINER_IMAGE_TAG",

	DebugPackagesKey: "CONTAINER_DEBUG_PACKAGES",
}

// supported distributions
var SupportedDistributions = []*core.SupportedDistribution{
	common.CommonSupportedDistribution,
	alpine.AlpineSupportedDistribution,
	debian.DebianSupportedDistribution,
}

// filesystem configuration
var FilesystemConfig = &core.FilesystemConfig{
	RootPath:    "/container",
	RunRootPath: "/run/container",

	EnvironmentFilesPrefix: ".env",
}

// services configuration
var ServicesConfig = &core.ServicesConfig{
	PriorityFilename: ".priority",
	DefaultPriority:  500,

	InstallFilename:   "install.sh",
	InstalledFilename: ".installed",
	StartupFilename:   "startup.sh",
	ProcessFilename:   "process.sh",
	FinishFilename:    "finish.sh",

	OptionalFilename: ".optional",
	DownloadFilename: "download.sh",
}

// generator configuration
var GeneratorConfig = &core.GeneratorConfig{
	TemplatesFilesSuffix: ".template",

	DockerfileTemplate:             "Dockerfile.template",
	DockerfileMultiprocessTemplate: "Dockerfile.multiprocess.template",

	ServicesTemplatesDir: "/services/service-name",

	EnvironmentTemplatesDir: "/environment",
}

// core configuration
var CoreConfig = &core.CoreConfig{
	Image:                  fmt.Sprintf("%v:%v", BuildImageName, BuildImageTag),
	SupportedDistributions: SupportedDistributions,

	EnvironmentConfig: CoreEnvironmentConfig,
	FilesystemConfig:  FilesystemConfig,
	ServicesConfig:    ServicesConfig,
	GeneratorConfig:   GeneratorConfig,
}
