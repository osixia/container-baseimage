package job

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"

	"github.com/osixia/container-baseimage/build/config"
	"github.com/osixia/container-baseimage/build/helpers"
	"github.com/osixia/container-baseimage/build/job"
)

type BuildFlags struct {
	dockerfile       string
	dockerfileTarget string

	version string
	latest  bool

	tagsSuffix string

	distributions []string
	arches        []string
}

var buildCmdFlags = &BuildFlags{}

var buildCmd = helpers.NewJobCmd("build", "Run build job", "b", job.Build, buildCmdFlags)

func init() {
	// flags
	buildCmd.Flags().SortFlags = false
	AddBuildFlags(buildCmd.Flags(), buildCmdFlags)
}

func AddBuildFlags(fs *pflag.FlagSet, bf *BuildFlags) {
	distributions := config.Distributions
	arches := arches()

	fs.StringVar(&bf.dockerfileTarget, "target", "", "dockerfile target\n")

	fs.StringVarP(&bf.version, "version", "v", config.DefaultVersion, "version")
	fs.BoolVarP(&bf.latest, "latest", "l", false, "tag as latest image\n")

	fs.StringVar(&bf.tagsSuffix, "tags-suffix", "", "tags suffix\n")

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

func (bf *BuildFlags) SetDockerfile(df string) {
	bf.dockerfile = df
}

func (bf *BuildFlags) ToJobOptions() any {

	return &job.BuildOptions{
		BuildImageOptions: job.BuildImageOptions{
			DockerfileTarget: bf.dockerfileTarget,

			Version:    bf.version,
			Latest:     bf.latest,
			TagsSuffix: bf.tagsSuffix,

			Platforms: filterPlatforms(bf.arches),
		},

		Dockerfile: bf.dockerfile,

		Images: filterImages(bf.distributions),
	}
}
