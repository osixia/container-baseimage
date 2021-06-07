package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	var DebianBullseyeImage = &config.Image{
		BaseImage:    "debian:bullseye-slim",
		Distribution: config.Debian,

		Name:        "osixia/baseimage",
		TagPrefixes: []string{"debian-bullseye", "debian-11"},
	}

	var Ubuntu2404Image = &config.Image{
		BaseImage:    "ubuntu:24.04",
		Distribution: config.Ubuntu,

		Name:        "osixia/baseimage",
		TagPrefixes: []string{"ubuntu-24.04", "ubuntu-noble", "ubuntu"},
	}

	var Ubuntu2204Image = &config.Image{
		BaseImage:    "ubuntu:22.04",
		Distribution: config.Ubuntu,

		Name:        "osixia/baseimage",
		TagPrefixes: []string{"ubuntu-22.04", "ubuntu-jammy"},
	}

	var Alpine323Image = &config.Image{
		BaseImage:    "alpine:3.23.3",
		Distribution: config.Alpine,

		Name:        "osixia/baseimage",
		TagPrefixes: []string{"alpine-3.23", "alpine-3", "alpine"},
	}

	var Alpine322Image = &config.Image{
		BaseImage:    "alpine:3.22.3",
		Distribution: config.Alpine,

		Name:        "osixia/baseimage",
		TagPrefixes: []string{"alpine-3.22"},
	}

	var Alpine321Image = &config.Image{
		BaseImage:    "alpine:3.21.6",
		Distribution: config.Alpine,

		Name:        "osixia/baseimage",
		TagPrefixes: []string{"alpine-3.21"},
	}

	config.Images = []*config.Image{
		DebianTrixieImage,
		DebianBookwormImage,
		DebianBullseyeImage,

		Ubuntu2404Image,
		Ubuntu2204Image,

		Alpine323Image,
		Alpine322Image,
		Alpine321Image,
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

	// custom job

	job.Export = func(ctx context.Context, client *dagger.Client, options any) ([]*job.BuildResult, error) {

		o := options.(*job.ExportOptions)

		builds, err := job.DefaultExport(ctx, client, o)
		if err != nil {
			return nil, err
		}

		if o.ExportDir == "" {
			return builds, nil
		}

		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		exportDir, err := filepath.Abs(filepath.Join(wd, o.ExportDir))
		if err != nil {
			return nil, err
		}

		// container to compress bin
		tar := client.Container().From(DebianTrixieImage.BaseImage)

		// create empty directory to put build outputs
		outputs := client.Directory()

		// export bin from default image in outputs directory
		for _, b := range builds {

			// export bin only from default image container with root
			if b.Image != config.DefaultImage {
				continue
			}

			for _, c := range b.Containers {

				p, err := c.Platform(ctx)
				if err != nil {
					return nil, err
				}
				ps := strings.ReplaceAll(string(p), "/", "_")

				// create dagger directory with compressed bin
				binName := fmt.Sprintf("container_%v", ps)
				tarName := fmt.Sprintf("%v.tar.gz", binName)

				tar, _ = tar.WithFile(binName, c.File("/usr/sbin/container")).Sync(ctx)
				tar = tar.WithExec([]string{"tar", "-czf", tarName, binName})

				outputs = outputs.WithFile(tarName, tar.File(tarName))

				b.Files = append(b.Files, filepath.Join(exportDir, tarName))
			}

			break
		}

		// export outputs directory on filesystem
		_, err = outputs.Export(ctx, exportDir)
		if err != nil {
			return nil, err
		}

		return builds, nil
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
