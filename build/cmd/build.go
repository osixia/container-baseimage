package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/build/config"
	"github.com/osixia/container-baseimage/build/job"
)

type buildFlags struct {
	dockerfile string

	version string
	latest  bool

	contributors string

	distributions []string
	arches        []string

	withNonroot bool
}

var buildCmdFlags = &buildFlags{}

var buildCmd = newStepCmd("build", "Run build job", "b", job.Build, buildCmdFlags)

func init() {
	// flags
	buildCmd.Flags().SortFlags = false
	addBuildFlags(buildCmd.Flags(), buildCmdFlags)
}

func addBuildFlags(fs *pflag.FlagSet, cf *buildFlags) {
	distributions := distributions()
	arches := arches()

	fs.StringSliceVarP(&cf.distributions, "distributions", "d", distributions, fmt.Sprintf("distributions to build, choices: %v", strings.Join(distributions, ", ")))
	fs.StringSliceVarP(&cf.arches, "arches", "a", arches, fmt.Sprintf("arches to build, choices: %v", strings.Join(arches, ", ")))
	fs.BoolVarP(&cf.withNonroot, "with-nonroot", "n", false, "build also nonroot containers")
	fs.BoolVarP(&cf.latest, "latest", "l", false, "tag as latest image")
	fs.StringVarP(&cf.contributors, "contributors", "c", "", "generated image contributors\n")
}

func distributions() []string {

	distributions := map[string]bool{}
	for _, img := range config.Images {
		distributions[img.Distribution] = true
	}

	distributionsStrings := make([]string, 0, len(distributions))
	for k := range distributions {
		distributionsStrings = append(distributionsStrings, k)
	}

	return distributionsStrings
}

func arches() []string {

	arches := map[string]bool{}
	for _, p := range config.Platforms {
		arches[p.GoArch] = true
	}

	archesStrings := make([]string, 0, len(arches))
	for k := range arches {
		archesStrings = append(archesStrings, k)
	}

	return archesStrings
}

func filterImages(distributions []string) []*config.Image {

	images := make([]*config.Image, 0)

	for _, d := range distributions {
		for _, ci := range config.Images {
			if ci.Distribution == d {
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
		BuildImageOptions: job.BuildImageOptions{
			Version: bf.version,
			Latest:  bf.latest,

			Contributors: bf.contributors,

			Platforms: filterPlatforms(bf.arches),
		},

		Dockerfile: bf.dockerfile,

		Images: filterImages(bf.distributions),

		WithNonroot: bf.withNonroot,
	}
}
