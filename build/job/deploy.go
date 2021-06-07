package job

import (
	"context"
	"fmt"
	"regexp"

	"dagger.io/dagger"

	"github.com/osixia/container-baseimage/build/config"
)

type DeployOptions struct {
	TestOptions

	Registry string
	Username string
	Password string

	DryRun bool
}

func (do *DeployOptions) Validate() (bool, error) {

	if do.Latest && !regexp.MustCompile(config.ReleaseTagRegex).MatchString(do.Version) {
		return false, fmt.Errorf("with latest set, version %v must be a tag that match %v", do.Version, config.ReleaseTagRegex)
	}

	return true, nil
}

func DefaultDeployJob(ctx context.Context, client *dagger.Client, options any) ([]*BuildResult, error) {

	o := options.(*DeployOptions)

	if _, err := o.Validate(); err != nil {
		return nil, err
	}

	builds, err := TestJob(ctx, client, &o.TestOptions)
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

			img := fmt.Sprintf("%v:%v", b.Image.Name, tag)

			c := client.Container()

			if o.Registry != "" {

				if o.Registry != "" {
					img = fmt.Sprintf("%v/%v", o.Registry, img)
				}

				if o.Username != "" {
					c = c.WithRegistryAuth(o.Registry, o.Username, client.SetSecret(o.Password, o.Password))
				}

			}

			fmt.Printf("pushing image %v ...\n", img)

			if o.DryRun {
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
