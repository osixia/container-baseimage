package main

import (
	"context"
	"os"

	"dagger.io/dagger"

	"github.com/osixia/container-baseimage/build/cmd"
	"github.com/osixia/container-baseimage/build/config"
	"github.com/osixia/container-baseimage/build/job"
)

func main() {

	// images

	var DebianTrixieImage = &config.Image{
		BaseImage:    "debian:trixie-slim",
		Distribution: config.Debian,

		Name:        "osixia/baseimage",
		TagPrefixes: []string{"debian-trixie", "debian-13", "debian"},
	}

	var DebianBookwormImage = &config.Image{
		BaseImage:    "debian:bookworm-slim",
		Distribution: config.Debian,

		Name:        "osixia/baseimage",
		TagPrefixes: []string{"debian-bookworm", "debian-12"},
	}

	var Ubuntu2604Image = &config.Image{
		BaseImage:    "ubuntu:26.04",
		Distribution: config.Ubuntu,

		Name:        "osixia/baseimage",
		TagPrefixes: []string{"ubuntu-26.04", "ubuntu-resolute-raccoon", "ubuntu-resolute", "ubuntu"},
	}

	var Ubuntu2404Image = &config.Image{
		BaseImage:    "ubuntu:24.04",
		Distribution: config.Ubuntu,

		Name:        "osixia/baseimage",
		TagPrefixes: []string{"ubuntu-24.04", "ubuntu-noble"},
	}

	var Alpine324Image = &config.Image{
		BaseImage:    "alpine:3.24",
		Distribution: config.Alpine,

		Name:        "osixia/baseimage",
		TagPrefixes: []string{"alpine-3.24", "alpine-3", "alpine"},
	}

	var Alpine323Image = &config.Image{
		BaseImage:    "alpine:3.23",
		Distribution: config.Alpine,

		Name:        "osixia/baseimage",
		TagPrefixes: []string{"alpine-3.23"},
	}

	var Alpine322Image = &config.Image{
		BaseImage:    "alpine:3.22",
		Distribution: config.Alpine,

		Name:        "osixia/baseimage",
		TagPrefixes: []string{"alpine-3.22"},
	}

	config.Images = []*config.Image{
		DebianTrixieImage,
		DebianBookwormImage,

		Ubuntu2604Image,
		Ubuntu2404Image,

		Alpine324Image,
		Alpine323Image,
		Alpine322Image,
	}

	config.DefaultImage = DebianTrixieImage

	for _, i := range config.Images {
		i.Description = "A container base image to build reliable single- or multi-process images quickly 🐳✨🌴"

		i.Url = "https://opensource.osixia.net/projects/container-images/baseimage/"
		i.Documentation = "https://opensource.osixia.net/projects/container-images/baseimage/"
		i.Source = "https://github.com/osixia/container-baseimage"

		i.Authors = "The osixia/container-baseimage maintainers 🐒✨🌴"
		i.Vendor = "Osixia"

		i.Licences = "MIT"
	}

	// github

	config.ProjectGithubRepo = &config.GithubRepo{
		Organization: "osixia",
		Project:      "container-baseimage",
	}

	// custom function

	job.BuildArgs = func(options *job.BuildImageOptions, image *config.Image, platform *config.Platform, tag string) []dagger.BuildArg {

		var buildArgs = job.DefaultBuildArgs(options, image, platform, tag)

		// common build args
		if options.Version != "" {
			versionArg := dagger.BuildArg{
				Name:  "VERSION",
				Value: options.Version,
			}
			buildArgs = append(buildArgs, versionArg)
		}

		// platform build args
		goarchArg := dagger.BuildArg{
			Name:  "GOARCH",
			Value: platform.GoArch,
		}

		// nonroot group args
		nonrootGroupIDArg := dagger.BuildArg{
			Name:  "NONROOT_GROUP_ID",
			Value: "65532",
		}
		nonrootGroupNameArg := dagger.BuildArg{
			Name:  "NONROOT_GROUP_NAME",
			Value: "nonroot",
		}

		// nonroot user args
		nonrootUserIDArg := dagger.BuildArg{
			Name:  "NONROOT_USER_ID",
			Value: "65532",
		}
		nonrootUserNameArg := dagger.BuildArg{
			Name:  "NONROOT_USER_NAME",
			Value: "nonroot",
		}

		return append(buildArgs, goarchArg, nonrootGroupIDArg, nonrootGroupNameArg, nonrootUserIDArg, nonrootUserNameArg)
	}

	// execute cmd

	mainCtx := context.Background()
	if err := cmd.Run(mainCtx); err != nil {
		os.Exit(1)
	}

}
