package job

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/osixia/container-baseimage/ci/config"
)

type BuildOptions struct {
	Dockerfile string

	Version      string
	Contributors string

	Images    []*config.Image
	Platforms []*config.Platform
}

type buildResult struct {
	Image      *config.Image
	Containers []*dagger.Container
}

func (bopt *BuildOptions) BuildImageTag(i *config.Image) string {
	return i.TagPrefixes[0] + "-" + bopt.Version
}

func build(ctx context.Context, client *dagger.Client, options *BuildOptions) ([]*buildResult, error) {

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	dp, err := filepath.Abs(filepath.Join(wd, options.Dockerfile))
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
		Exclude: []string{".git/", ".github/", "bin/", "ci/", "docs/", "*.md", ".dockerignore", ".gitignore", "Dockerfile"},
	}

	contextDir := client.Host().Directory(dockerfileDir, hostDirectoryOpts)
	dockerfile := client.Host().File(dp)
	workspace := contextDir.WithFile(dockerfileName, dockerfile)

	entries, err := workspace.Entries(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("working dir content: %v\n", entries)

	// container to compress bin
	tar := client.Container().From("ubuntu:latest")

	// create empty directory to put build outputs
	outputs := client.Directory()

	builds := make([]*buildResult, 0, len(options.Images))
	for _, image := range options.Images {

		b := &buildResult{
			Image:      image,
			Containers: make([]*dagger.Container, 0, len(options.Platforms)),
		}

		for _, platform := range options.Platforms {

			opt := dagger.ContainerBuildOpts{
				BuildArgs:  buildArgs(options, image, platform),
				Dockerfile: dockerfileName,
			}

			c, err := client.Container(dagger.ContainerOpts{Platform: platform.Name}).Build(workspace, opt).Sync(ctx)
			if err != nil {
				return nil, err
			}

			binName := fmt.Sprintf("container-baseimage_%s_%s", options.BuildImageTag(image), platform.GoArch)
			tarName := fmt.Sprintf("%v.tar.gz", binName)

			tar, _ = tar.WithFile(binName, c.File("/usr/sbin/container-baseimage")).Sync(ctx)
			tar = tar.WithExec([]string{"tar", "-czf", tarName, binName})

			outputs = outputs.WithFile(tarName, tar.File(tarName))

			b.Containers = append(b.Containers, c)
		}

		_, err = outputs.Export(ctx, filepath.Join(dockerfileDir, "/bin"))
		if err != nil {
			return nil, err
		}

		builds = append(builds, b)
	}

	return builds, nil
}

func buildArgs(options *BuildOptions, image *config.Image, platform *config.Platform) []dagger.BuildArg {

	buildArgs := []dagger.BuildArg{}

	// common build args
	if options.Version != "" {
		buildVersionArg := dagger.BuildArg{
			Name:  "BUILD_VERSION",
			Value: options.Version,
		}
		buildArgs = append(buildArgs, buildVersionArg)
	}

	if options.Contributors != "" {
		buildContributorsArg := dagger.BuildArg{
			Name:  "BUILD_CONTRIBUTORS",
			Value: options.Contributors,
		}
		buildArgs = append(buildArgs, buildContributorsArg)
	}

	// image build args
	buildRootImageNameArg := dagger.BuildArg{
		Name:  "ROOT_IMAGE_NAME",
		Value: image.RootImageName,
	}

	buildRootImageTagArg := dagger.BuildArg{
		Name:  "ROOT_IMAGE_TAG",
		Value: image.RootImageTag,
	}

	buildImageNameArg := dagger.BuildArg{
		Name:  "BUILD_IMAGE_NAME",
		Value: image.BuildImageName,
	}

	buildImageTagArg := dagger.BuildArg{
		Name:  "BUILD_IMAGE_TAG",
		Value: options.BuildImageTag(image),
	}

	// platform build args
	goarchArg := dagger.BuildArg{
		Name:  "GOARCH",
		Value: platform.GoArch,
	}

	return append(buildArgs, buildRootImageNameArg, buildRootImageTagArg, buildImageNameArg, buildImageTagArg, goarchArg)
}
