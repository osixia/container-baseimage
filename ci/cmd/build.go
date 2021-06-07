package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/ci/config"
	"github.com/osixia/container-baseimage/ci/job"
)

type buildFlags struct {
	dockerfile string

	version      string
	contributors string

	images []string
	arches []string
}

var buildCmdFlags = &buildFlags{}

var buildCmd = newStepCmd("build", "Run build job", "b", job.Build, buildCmdFlags)

func init() {
	// flags
	buildCmd.Flags().SortFlags = false
	addBuildFlags(buildCmd.Flags(), buildCmdFlags)
}

func addBuildFlags(fs *pflag.FlagSet, cf *buildFlags) {
	images := images()
	arches := arches()

	fs.StringSliceVarP(&cf.images, "images", "i", images, fmt.Sprintf("images to build, choices: %v", strings.Join(images, ", ")))
	fs.StringSliceVarP(&cf.arches, "arches", "a", arches, fmt.Sprintf("arches to build, choices: %v", strings.Join(arches, ", ")))
	fs.StringVarP(&cf.contributors, "contributors", "c", "", "generated image contributors\n")
}

func images() []string {

	rootImgs := map[string]bool{}
	for _, img := range config.Images {
		rootImgs[img.RootImageName] = true
	}

	var rootImgsStrings []string
	for k := range rootImgs {
		rootImgsStrings = append(rootImgsStrings, k)
	}

	return rootImgsStrings
}

func arches() []string {

	arches := map[string]bool{}
	for _, p := range config.Platforms {
		arches[p.GoArch] = true
	}

	var archesStrings []string
	for k := range arches {
		archesStrings = append(archesStrings, k)
	}

	return archesStrings
}

func filterImages(imgs []string) []*config.Image {

	images := make([]*config.Image, 0, len(imgs))

	for _, i := range imgs {
		for _, ci := range config.Images {
			if ci.RootImageName == i {
				images = append(images, ci)
			}
		}
	}

	return images
}

func filterPlatforms(pfs []string) []*config.Platform {

	platforms := make([]*config.Platform, 0, len(pfs))

	for _, pf := range pfs {
		for _, cpf := range config.Platforms {
			if cpf.GoArch == pf {
				platforms = append(platforms, cpf)
			}
		}
	}

	return platforms
}

func (bf *buildFlags) SetDockerfile(d string) {
	bf.dockerfile = d
}

func (bf *buildFlags) SetVersion(v string) {
	bf.version = v
}

func (bf *buildFlags) toJobOptions() interface{} {

	return &job.BuildOptions{
		Dockerfile: bf.dockerfile,

		Version:      bf.version,
		Contributors: bf.contributors,

		Images:    filterImages(bf.images),
		Platforms: filterPlatforms(bf.arches),
	}
}
