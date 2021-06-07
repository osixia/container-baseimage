package job

import (
	"context"
	"fmt"
	"regexp"

	"dagger.io/dagger"
)

type DeployOptions struct {
	TestOptions

	Registry string
	Username string
	Password string

	DryRun bool
}

func (do *DeployOptions) Validate() error {

	if do.Latest && !regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`).MatchString(do.BuildOptions.Version) {
		return fmt.Errorf("error: with latest set, version must be a tag formated like x.y.z with x, y and z numbers")
	}

	return nil
}

func deploy(ctx context.Context, client *dagger.Client, options *DeployOptions) ([]*BuildResult, error) {

	if err := options.Validate(); err != nil {
		return nil, err
	}

	builds, err := test(ctx, client, &options.TestOptions)
	if err != nil {
		return nil, err
	}

	for _, b := range builds {

		publishOptions := dagger.ContainerPublishOpts{
			PlatformVariants: b.Containers,
			// Some registries may require explicit use of docker mediatypes
			// rather than the default OCI mediatypes
			// MediaTypes: dagger.Dockermediatypes,
		}

		for _, tag := range b.Tags {

			img := fmt.Sprintf("%v:%v", b.Image.ImageName, tag)

			c := client.Container()

			if options.Registry != "" {

				if options.Registry != "" {
					img = fmt.Sprintf("%v/%v", options.Registry, img)
				}

				if options.Username != "" {
					c = c.WithRegistryAuth(options.Registry, options.Username, client.SetSecret(options.Password, options.Password))
				}

			}

			fmt.Printf("pushing image %v ...\n", img)

			if options.DryRun {
				continue
			}

			digest, err := c.Publish(ctx, img, publishOptions)
			if err != nil {
				return nil, err
			}
			fmt.Printf("image %v pushed with digest %v\n", img, digest)
		}
	}

	return builds, nil
}
