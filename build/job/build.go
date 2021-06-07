package job

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"dagger.io/dagger"

	"github.com/osixia/container-baseimage/build/config"
)

type BuildOptions struct {
	BuildImageOptions

	Dockerfile string

	Images []*config.Image

	WithNonroot bool
}

type BuildImageOptions struct {
	Version string
	Latest  bool

	Contributors string

	Platforms []*config.Platform
}

type BuildResult struct {
	Image      *config.Image
	NonRoot    bool
	Containers []*dagger.Container
	Tags       []string

	Files []string
}

func build(ctx context.Context, client *dagger.Client, options *BuildOptions) ([]*BuildResult, error) {

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
		Exclude: []string{".git/", ".github/", "bin/", "build/", "docs/", "*.md", ".dockerignore", ".gitignore", "Dockerfile"},
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
	for _, image := range options.Images {

		b, err := buildImage(ctx, workspace, dockerfileName, &options.BuildImageOptions, image, false)
		if err != nil {
			return nil, err
		}

		builds = append(builds, b)

		if options.WithNonroot {
			b, err := buildImage(ctx, workspace, dockerfileName, &options.BuildImageOptions, image, true)
			if err != nil {
				return nil, err
			}

			builds = append(builds, b)
		}
	}

	// container to compress bin
	tar := client.Container().From("ubuntu:latest")

	// create empty directory to put build outputs
	outputs := client.Directory()
	binDir := filepath.Join(dockerfileDir, "bin")

	// export bin from default image in outputs directory
	for _, b := range builds {

		// export bin only from default image container with root
		if b.Image != config.DefaultImage || b.NonRoot {
			continue
		}

		for _, c := range b.Containers {

			p, err := c.Platform(ctx)
			if err != nil {
				return nil, err
			}
			ps := strings.Replace(string(p), "/", "_", -1)

			// create dagger directory with compressed bin
			binName := fmt.Sprintf("container_%v", ps)
			tarName := fmt.Sprintf("%v.tar.gz", binName)

			tar, _ = tar.WithFile(binName, c.File("/usr/sbin/container")).Sync(ctx)
			tar = tar.WithExec([]string{"tar", "-czf", tarName, binName})

			outputs = outputs.WithFile(tarName, tar.File(tarName))

			b.Files = append(b.Files, filepath.Join(binDir, tarName))
		}
	}

	// export outputs directory on filesystem
	_, err = outputs.Export(ctx, binDir)
	if err != nil {
		return nil, err
	}

	return builds, nil
}

func buildImage(ctx context.Context, workspace *dagger.Directory, dockerfile string, options *BuildImageOptions, image *config.Image, nonroot bool) (*BuildResult, error) {

	b := &BuildResult{
		Image:      image,
		Containers: make([]*dagger.Container, 0, len(options.Platforms)),
		NonRoot:    nonroot,

		Tags: buildTags(image, options.Version, options.Latest, nonroot),
	}

	for _, platform := range options.Platforms {

		cbo := dagger.DirectoryDockerBuildOpts{
			BuildArgs:  buildArgs(image, options.Version, options.Contributors, platform, b.LongestTag()),
			Dockerfile: dockerfile,
			Platform:   platform.Name,
		}

		c, err := workspace.DockerBuild(cbo).Sync(ctx)

		if err != nil {
			return nil, err
		}

		if nonroot {
			if image.Distribution == "alpine" {
				c = c.WithExec([]string{"adduser", "-D", config.NonrootUser.Name, "-u", strconv.Itoa(config.NonrootUser.ID), "-g", config.NonrootGroup.Name, "-s", "/sbin/nologin"})
			} else if slices.Contains([]string{"debian", "ubuntu"}, image.Distribution) {
				c = c.WithExec([]string{"groupadd", config.NonrootGroup.Name, "-g", strconv.Itoa(config.NonrootGroup.ID)}).
					WithExec([]string{"useradd", "-m", config.NonrootUser.Name, "-u", strconv.Itoa(config.NonrootUser.ID), "-g", strconv.Itoa(config.NonrootGroup.ID), "-s", "/sbin/nologin"})
			} else {
				panic(fmt.Sprintf("error: root image %v not supported", image.Distribution))
			}

			c = c.WithUser(config.NonrootUser.Name)
		}

		// add oci labels
		c = c.WithLabel("org.opencontainers.image.title", image.BuildImageName).
			WithLabel("org.opencontainers.image.version", b.LongestTag()).
			WithLabel("org.opencontainers.image.created", time.Now().String()).
			WithLabel("org.opencontainers.image.source", fmt.Sprintf("https://github.com/%v/%v", config.ProjectGithubRepo.Organization, config.ProjectGithubRepo.Project)).
			WithLabel("org.opencontainers.image.licenses", "MIT")

		c, err = c.Sync(ctx)
		if err != nil {
			return nil, err
		}

		b.Containers = append(b.Containers, c)
	}

	return b, nil
}

func buildArgs(image *config.Image, version string, contributors string, platform *config.Platform, tag string) []dagger.BuildArg {

	var buildArgs []dagger.BuildArg

	// common build args
	if version != "" {
		buildVersionArg := dagger.BuildArg{
			Name:  "BUILD_VERSION",
			Value: version,
		}
		buildArgs = append(buildArgs, buildVersionArg)
	}

	if contributors != "" {
		buildContributorsArg := dagger.BuildArg{
			Name:  "BUILD_CONTRIBUTORS",
			Value: contributors,
		}
		buildArgs = append(buildArgs, buildContributorsArg)
	}

	// image build args
	buildRootImageArg := dagger.BuildArg{
		Name:  "ROOT_IMAGE",
		Value: image.RootImage,
	}

	buildImageNameArg := dagger.BuildArg{
		Name:  "BUILD_IMAGE_NAME",
		Value: image.BuildImageName,
	}

	buildImageTagArg := dagger.BuildArg{
		Name:  "BUILD_IMAGE_TAG",
		Value: tag,
	}

	// platform build args
	goarchArg := dagger.BuildArg{
		Name:  "GOARCH",
		Value: platform.GoArch,
	}

	return append(buildArgs, buildRootImageArg, buildImageNameArg, buildImageTagArg, goarchArg)
}

func buildTags(image *config.Image, version string, latest bool, nonroot bool) []string {
	var tags []string

	v := strings.Split(version, ".")

	// default image tags
	if image == config.DefaultImage {

		// version x.y.z
		tags = append(tags, version)

		// latest release
		if latest {

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
		tags = append(tags, fmt.Sprintf("%v-%v", tagPrefix, version))

		// latest release
		if latest {

			// prefix + version x.y
			tags = append(tags, fmt.Sprintf("%v-%v.%v", tagPrefix, v[0], v[1]))

			// prefix + version x
			tags = append(tags, fmt.Sprintf("%v-%v", tagPrefix, v[0]))

			// prefix only
			tags = append(tags, fmt.Sprintf("%v", tagPrefix))
		}
	}

	// add nonroot suffix
	if nonroot {
		for i := range tags {
			tags[i] += "-nonroot"
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
