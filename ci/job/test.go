package job

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
)

// version
// no process (check env files)
// single process (check env files)
// multiprocess (check env files)
// run only one service
// no startup / process / finish
// skip env files
// pre startup / process / finish file
// restart processes

// kill all ?

// debug
// install packages

type TestOptions struct {
	BuildOptions
}

func test(ctx context.Context, client *dagger.Client, options *TestOptions) ([]*buildResult, error) {

	builds, err := build(ctx, client, &options.BuildOptions)
	if err != nil {
		return nil, err
	}

	for _, b := range builds {

		expectedImageVersion := fmt.Sprintf("%v:%v", b.Image.BuildImageName, options.BuildImageTag(b.Image))

		for _, i := range b.Containers {

			imgVersion, err := i.WithExec([]string{"--version"}).Stdout(ctx)
			if err != nil {
				return nil, err
			}

			if strings.TrimSuffix(imgVersion, "\n") != expectedImageVersion {
				return nil, fmt.Errorf("error: image version is %s expected image version is %s", imgVersion, expectedImageVersion)
			}
		}

	}

	return builds, nil
}
