package job

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"

	"github.com/osixia/container-baseimage/build/config"
)

type JobFunc func(ctx context.Context, client *dagger.Client, options any) ([]*BuildResult, error)

type BuildTagsFunc func(image *config.Image, options *BuildImageOptions, suffix string) []string
type BuildArgsFunc func(options *BuildImageOptions, image *config.Image, platform *config.Platform, tag string) []dagger.BuildArg

var Build JobFunc = DefaultBuild
var Export JobFunc = DefaultExport
var Test JobFunc = DefaultTest
var Deploy JobFunc = DefaultDeploy
var Publish JobFunc = DefaultPublish

var BuildTags BuildTagsFunc = DefaultBuildTags
var BuildArgs BuildArgsFunc = DefaultBuildArgs

func Run(ctx context.Context, f JobFunc, options any) ([]*BuildResult, error) {

	// dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Print(err.Error())
		}
	}()

	return f(ctx, client, options)
}
