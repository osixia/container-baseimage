package job

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"dagger.io/dagger"

	"github.com/osixia/container-baseimage/ci/config"
)

type DeployOptions struct {
	TestOptions

	Latest bool
	DryRun bool
}

func (do *DeployOptions) Validate() error {

	if do.Latest && !regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(do.BuildOptions.Version) {
		return fmt.Errorf("error: with latest set, version must be a tag formated like x.y.z with x, y and z numbers")
	}

	return nil
}

func deploy(ctx context.Context, client *dagger.Client, options *DeployOptions) ([]*buildResult, error) {

	if err := options.Validate(); err != nil {
		return nil, err
	}

	builds, err := test(ctx, client, &options.TestOptions)
	if err != nil {
		return nil, err
	}

	for _, b := range builds {

		imgs := []string{}

		v := strings.Split(options.Version, ".")

		// default image tags
		if b.Image == config.DefaultImage {

			// version x.y.z
			imgs = append(imgs, fmt.Sprintf("%v:%v", b.Image.BuildImageName, options.Version))

			// latest release
			if options.Latest {

				// latest
				imgs = append(imgs, fmt.Sprintf("%v:latest", b.Image.BuildImageName))

				// version x.y
				imgs = append(imgs, fmt.Sprintf("%v:%v.%v", b.Image.BuildImageName, v[0], v[1]))

				// version x
				imgs = append(imgs, fmt.Sprintf("%v:%v", b.Image.BuildImageName, v[0]))
			}

		}

		// regular image tags
		for _, tagPrefix := range b.Image.TagPrefixes {

			// prefix + version x.y.z
			imgs = append(imgs, fmt.Sprintf("%v:%v-%v", b.Image.BuildImageName, tagPrefix, options.Version))

			// latest release
			if options.Latest {

				// prefix + version x.y
				imgs = append(imgs, fmt.Sprintf("%v:%v-%v.%v", b.Image.BuildImageName, tagPrefix, v[0], v[1]))

				// prefix + version x
				imgs = append(imgs, fmt.Sprintf("%v:%v-%v", b.Image.BuildImageName, tagPrefix, v[0]))

				// prefix only
				imgs = append(imgs, fmt.Sprintf("%v:%v", b.Image.BuildImageName, tagPrefix))
			}
		}

		publishOptions := dagger.ContainerPublishOpts{
			PlatformVariants: b.Containers,
			// Some registries may require explicit use of docker mediatypes
			// rather than the default OCI mediatypes
			// MediaTypes: dagger.Dockermediatypes,
		}

		for _, img := range imgs {
			fmt.Printf("pushing image %v ...\n", img)

			if options.DryRun {
				continue
			}

			digest, err := client.Container().Publish(ctx, img, publishOptions)
			if err != nil {
				return nil, err
			}
			fmt.Printf("image %v pushed with digest %v\n", img, digest)
		}
	}

	return builds, nil
}
