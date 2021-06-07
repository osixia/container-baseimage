package job

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dagger.io/dagger"

	"github.com/osixia/container-baseimage/build/config"
)

type BuildOptions struct {
	BuildImageOptions

	Dockerfile string
	Images     []*config.Image
}

type BuildImageOptions struct {
	DockerfileTarget string

	Version    string
	Latest     bool
	TagsSuffix string

	Platforms []*config.Platform
}

type BuildResult struct {
	Image      *config.Image
	Containers []*dagger.Container
	Tags       []string

	Files []string
}

func DefaultBuild(ctx context.Context, client *dagger.Client, options any) ([]*BuildResult, error) {

	o := options.(*BuildOptions)

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	dp, err := filepath.Abs(filepath.Join(wd, o.Dockerfile))
	if err != nil {
		return nil, err
	}

	// check dockerfile exists
	if _, err := os.Stat(dp); err != nil {
		return nil, err
	}

	dockerfileDir := filepath.Dir(dp)
	dockerfileName := filepath.Base(dp)

	hostDirectoryOpts := dagger.HostDirectoryOpts{
		Exclude: config.ExcludedBuildPaths,
	}

	contextDir := client.Host().Directory(dockerfileDir, hostDirectoryOpts)
	dockerfile := client.Host().File(dp)
	workspace := contextDir.WithFile(dockerfileName, dockerfile)

	entries, err := workspace.Entries(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("working dir content: %v\n", entries)

	var builds []*BuildResult
	for _, image := range o.Images {

		if _, err := image.Validate(); err != nil {
			return nil, err
		}

		b, err := BuildImage(ctx, workspace, dockerfileName, &o.BuildImageOptions, image)
		if err != nil {
			return nil, err
		}

		builds = append(builds, b)
	}

	return builds, nil
}

func BuildImage(ctx context.Context, workspace *dagger.Directory, dockerfile string, options *BuildImageOptions, image *config.Image) (*BuildResult, error) {

	b := &BuildResult{
		Image:      image,
		Containers: make([]*dagger.Container, 0, len(options.Platforms)),
		Tags:       BuildTags(image, options, options.TagsSuffix),
	}

	for _, platform := range options.Platforms {

		cbo := dagger.DirectoryDockerBuildOpts{
			BuildArgs:  BuildArgs(options, image, platform, b.LongestTag()),
			Dockerfile: dockerfile,
			Target:     options.DockerfileTarget,
			Platform:   platform.Name,
		}

		c, err := workspace.DockerBuild(cbo).Sync(ctx)

		if err != nil {
			return nil, err
		}

		// add oci labels
		c = c.WithLabel("org.opencontainers.image.title", image.Name).
			WithLabel("org.opencontainers.image.version", options.Version).
			WithLabel("org.opencontainers.image.ref.name", b.LongestTag()).
			WithLabel("org.opencontainers.image.base.name", image.BaseImage).
			WithLabel("org.opencontainers.image.created", time.Now().String())

		if image.Description != "" {
			c = c.WithLabel("org.opencontainers.image.description", image.Description)
		}

		if image.Url != "" {
			c = c.WithLabel("org.opencontainers.image.url", image.Url)
		}

		if image.Documentation != "" {
			c = c.WithLabel("org.opencontainers.image.documentation", image.Documentation)
		}

		if image.Source != "" {
			c = c.WithLabel("org.opencontainers.image.source", image.Source)
		}

		if image.Authors != "" {
			c = c.WithLabel("org.opencontainers.image.authors", image.Authors)
		}

		if image.Vendor != "" {
			c = c.WithLabel("org.opencontainers.image.vendor", image.Vendor)
		}

		if image.Licences != "" {
			c = c.WithLabel("org.opencontainers.image.licenses", image.Licences)
		}

		c, err = c.Sync(ctx)
		if err != nil {
			return nil, err
		}

		b.Containers = append(b.Containers, c)
	}

	return b, nil
}

func DefaultBuildArgs(options *BuildImageOptions, image *config.Image, platform *config.Platform, tag string) []dagger.BuildArg {

	var buildArgs []dagger.BuildArg

	// image build args
	baseImageArg := dagger.BuildArg{
		Name:  "BASE_IMAGE",
		Value: image.BaseImage,
	}

	imageArg := dagger.BuildArg{
		Name:  "IMAGE",
		Value: fmt.Sprintf("%v:%v", image.Name, tag),
	}

	return append(buildArgs, baseImageArg, imageArg)
}

func DefaultBuildTags(image *config.Image, options *BuildImageOptions, suffix string) []string {
	var tags []string

	v := strings.Split(options.Version, ".")

	// default image tags
	if image == config.DefaultImage {

		// version x.y.z
		tags = append(tags, options.Version)

		// latest release
		if options.Latest {

			// latest
			tags = append(tags, "latest")

			// version x.y
			tags = append(tags, fmt.Sprintf("%v.%v", v[0], v[1]))

			// version x
			tags = append(tags, v[0])
		}
	}

	// regular image tags
	for _, tagPrefix := range image.TagPrefixes {

		// prefix + version x.y.z
		tags = append(tags, fmt.Sprintf("%v-%v", tagPrefix, options.Version))

		// latest release
		if options.Latest {

			// prefix + version x.y
			tags = append(tags, fmt.Sprintf("%v-%v.%v", tagPrefix, v[0], v[1]))

			// prefix + version x
			tags = append(tags, fmt.Sprintf("%v-%v", tagPrefix, v[0]))

			// prefix only
			tags = append(tags, fmt.Sprintf("%v", tagPrefix))
		}
	}

	// add suffix
	if suffix != "" {
		for i := range tags {
			tags[i] += suffix
		}
	}

	return tags
}

func (br *BuildResult) LongestTag() string {

	if len(br.Tags) == 0 {
		return ""
	}

	longest := br.Tags[0]
	for _, str := range br.Tags {
		if len(str) > len(longest) {
			longest = str
		}
	}

	return longest
}
