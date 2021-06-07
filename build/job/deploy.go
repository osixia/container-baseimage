package job

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"dagger.io/dagger"

	"github.com/osixia/container-baseimage/build/config"
)

type DeployOptions struct {
	TestOptions

	Registry string
	Username string
	Password string

	DigestsFile string

	DryRun bool
}

func (do *DeployOptions) Validate() (bool, error) {

	if do.Latest && !regexp.MustCompile(config.ReleaseTagRegex).MatchString(do.Version) {
		return false, fmt.Errorf("with latest set, version %v must be a tag that match %v", do.Version, config.ReleaseTagRegex)
	}

	return true, nil
}

func DefaultDeploy(ctx context.Context, client *dagger.Client, options any) ([]*BuildResult, error) {

	o := options.(*DeployOptions)

	if _, err := o.Validate(); err != nil {
		return nil, err
	}

	builds, err := Test(ctx, client, &o.TestOptions)
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

			if o.DigestsFile != "" {
				f, err := os.OpenFile(o.DigestsFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return nil, fmt.Errorf("opening cosign output file: %w", err)
				}
				_, writeErr := fmt.Fprintf(f, "%v\n", digest)
				if closeErr := f.Close(); closeErr != nil && writeErr == nil {
					writeErr = closeErr
				}
				if writeErr != nil {
					return nil, fmt.Errorf("writing to cosign output file: %w", writeErr)
				}
			}
		}
	}

	return builds, nil
}
