package job

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
)

// version

type TestOptions struct {
	ExportOptions
}

func DefaultTestJob(ctx context.Context, client *dagger.Client, options any) ([]*BuildResult, error) {

	o := options.(*TestOptions)

	builds, err := ExportJob(ctx, client, &o.ExportOptions)
	if err != nil {
		return nil, err
	}

	for _, b := range builds {

		expectedImageVersion := fmt.Sprintf("%v:%v", b.Image.Name, b.LongestTag())

		for _, i := range b.Containers {

			imgVersion, err := i.WithExec([]string{"--version"}, dagger.ContainerWithExecOpts{UseEntrypoint: true}).Stdout(ctx)
			if err != nil {
				return nil, err
			}

			if strings.TrimSuffix(imgVersion, "\n") != expectedImageVersion {
				return nil, fmt.Errorf("error: image version is %v expected image version is %v", imgVersion, expectedImageVersion)
			}
		}
	}

	return builds, nil
}
