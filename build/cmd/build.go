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

func addBuildFlags(fs *pflag.FlagSet, bf *buildFlags) {
	distributions := config.Distributions
	arches := arches()

	fs.BoolVarP(&bf.latest, "latest", "l", false, "tag as latest image\n")

	fs.BoolVarP(&bf.withNonroot, "with-nonroot", "n", false, "build also nonroot containers\n")

	fs.StringSliceVarP(&bf.distributions, "distributions", "d", distributions, fmt.Sprintf("distributions to build, choices: %v", strings.Join(distributions, ", ")))
	fs.StringSliceVarP(&bf.arches, "arches", "a", arches, fmt.Sprintf("arches to build, choices: %v", strings.Join(arches, ", ")))
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

func (bf *buildFlags) toJobOptions() any {

	return &job.BuildOptions{
		BuildImageOptions: job.BuildImageOptions{
			Version: bf.version,
			Latest:  bf.latest,

			Platforms: filterPlatforms(bf.arches),
		},

		Dockerfile: bf.dockerfile,

		Images: filterImages(bf.distributions),

		WithNonroot: bf.withNonroot,
	}
}
